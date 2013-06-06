/*
Copyright (c) 2013 Leandro López

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

/*
gedis server - Generic Redis server implementation

This package allows to create servers that can talk to clients by
using the Redis protocol

Redis protocol: http://redis.io/topics/protocol

As an example, the following implements a very simple Redis server
that only responds to the PING command:

    package main

    import (
    	"fmt"
    	gedis "github.com/inkel/gedis/server"
    	"os"
    	"os/signal"
    )

    func main() {
    	c := make(chan os.Signal, 1)
    	signal.Notify(c, os.Interrupt, os.Kill)

    	s, err := gedis.NewServer("tcp", ":10003")
    	if err != nil {
    		panic(err)
    	}
    	defer s.Close()

    	pong := []byte("+PONG\r\n")
    	earg := []byte("-ERR wrong number of arguments for 'ping' command\r\n")

    	s.Handle("PING", func(c *gedis.Client, args []string) error {
    		if len(args) != 0 {
    			c.Write(earg)
    			return nil
    		} else {
    			c.Write(pong)
    		}

    		return nil
    	})

    	go s.Loop()

    	sig := <-c

    	fmt.Println("Bye!", sig)
    }
*/
package server

import (
	"fmt"
	"github.com/inkel/gedis"
	"io"
	"net"
	"strings"
)

// Interface for reading from Redis clients
type Reader interface {
	Read([]byte) (int, error)
}

// Read a bulk as defined in the Redis protocol
func readBulk(r Reader) (bs []byte, err error) {
	var b byte

	b, err = readByte(r)
	if err != nil {
		return bs, err
	} else if b != '$' {
		return bs, gedis.NewParseError("Invalid first character")
	}

	n, err := gedis.ReadNumber(r)
	if err != nil {
		return bs, err
	}

	bs = make([]byte, n)

	_, err = r.Read(bs)
	if err != nil {
		return bs, err
	}

	crlf := make([]byte, 2)

	if _, err = r.Read(crlf); err != nil {
		return bs, err
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return bs, gedis.NewParseError("Invalid EOL")
	}

	return
}

// Reads the next byte in Reader
func readByte(r Reader) (byte, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	return b[0], err
}

// Read a multi-bulk request from a Redis client
//
// Redis client can only send multi-bulk requests to a Redis
// server. In truth they can also send an inline request, however that
// is currently not covered by this implementation
func Read(r Reader) (res [][]byte, err error) {
	var b byte

	b, err = readByte(r)
	if err != nil {
		return
	}

	if b != '*' {
		return res, gedis.NewParseError("Invalid first character")
	} else {
		n, err := gedis.ReadNumber(r)
		if err != nil {
			return res, err
		}

		res = make([][]byte, n)

		for i := int64(0); i < n; i++ {
			res[i], err = readBulk(r)
			if err != nil {
				return res, err
			}
		}
	}

	return res, err
}

// Holds pointers to the current Server and client net.Conn
type Client struct {
	server *Server
	conn   *net.Conn
}

// Disconnects a client
func (c *Client) Close() {
	conn := *c.conn
	conn.Close()
}

// Read from the client, parsing the input with the Redis protocol
func (c *Client) Read() ([][]byte, error) {
	return Read(*c.conn)
}

// Send a sequence of bytes to a client
func (c *Client) Write(bytes []byte) (int, error) {
	conn := *c.conn
	return conn.Write(bytes)
}

// Sends an error to the client, formatted accordingly to the Redis
// protocol
func (c *Client) Error(err error) (int, error) {
	return c.Write(gedis.WriteError(err))
}

// Sends a string formatted as an error to the client
func (c *Client) Errorf(format string, args ...interface{}) (int, error) {
	return c.Error(fmt.Errorf(format, args...))
}

// Sends a status response, formatted accordingly to the Redis
// protocol
func (c *Client) Status(status string) (int, error) {
	return c.Write(gedis.WriteStatus(status))
}

// Signature that command handler functions must have
type Handler func(c *Client, args [][]byte) error

// Structure to hold the necessary information to run a generic Redis
// server
type Server struct {
	ln       net.Listener
	handlers map[string]Handler
}

// Returns a new Server that listen in the specified network address
func NewServer(network, address string) (s Server, err error) {
	s.handlers = make(map[string]Handler)
	s.ln, err = net.Listen(network, address)
	return
}

// Closes a Redis server and stop processing
func (s *Server) Close() error {
	return s.ln.Close()
}

// Add a command handler
//
// Note that this function does not validate that the command is a
// valid Redis command, nor that the command hasn't already a handler.
func (s *Server) Handle(cmd string, handler Handler) {
	cmd = strings.ToUpper(cmd)
	s.handlers[cmd] = handler
}

// Goroutine to process data from a Client
func (s *Server) process(c *Client) {
	defer c.Close()

	for {
		in, err := c.Read()
		if err != nil {
			if err != io.EOF {
				c.Error(err)
			}
			return
		}

		cmd := strings.ToUpper(string(in[0]))

		if fn := s.handlers[cmd]; fn != nil {
			err = fn(c, in[1:])

			if err != nil {
				fmt.Printf("Unexpected error while processing connection: %v\n", err)
			}
		} else {
			c.Errorf("Unrecognized command '%s'", in[0])
		}
	}
}

// Main event loop for Redis clients
func (s *Server) Loop() {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			fmt.Printf("Error while accepting a connection: %v\n", err)
			continue
		}
		client := &Client{s, &c}
		go s.process(client)
	}
}
