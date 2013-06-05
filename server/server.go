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
that only respond to the PING command:

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
	"bytes"
	"fmt"
	"github.com/inkel/gedis"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

// Interface for reading from Redis clients
type Reader interface {
	Read([]byte) (int, error)
}

// Read the length of a bulk or multi-bulk block
func readLength(buf *bytes.Buffer) (n int64, err error) {
	sn, err := buf.ReadString('\r')

	if err != nil {
		return -1, err
	}

	b, err := buf.ReadByte()
	if err != nil {
		return -1, err
	} else if b != '\n' {
		return -1, fmt.Errorf("Invalid EOL: %q", []byte{'\r', b})
	}

	if n < 0 {
		return -1, fmt.Errorf("Negative length: %d", n)
	}

	return strconv.ParseInt(sn[:len(sn)-1], 10, 64)
}

// Read a bulk as defined in the Redis protocol
func readBulk(buf *bytes.Buffer) (bs []byte, err error) {
	var b byte

	b, err = buf.ReadByte()
	if err != nil {
		return bs, err
	} else if b != '$' {
		return bs, fmt.Errorf("Invalid first character: %q", b)
	}

	n, err := readLength(buf)
	if err != nil {
		return bs, err
	}

	bs = make([]byte, n)

	_, err = buf.Read(bs)
	if err != nil {
		return bs, err
	}

	crlf := make([]byte, 2)

	if _, err = buf.Read(crlf); err != nil {
		return bs, err
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return bs, fmt.Errorf("Invalid EOL: %q", crlf)
	}

	return
}

// Read a multi-bulk request from a Redis client
//
// Redis client can only send multi-bulk requests to a Redis
// server. In truth they can also send an inline request, however that
// is currently not covered by this implementation
func Read(r Reader) (res [][]byte, err error) {
	var b byte

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)

	b, err = buf.ReadByte()
	if err != nil {
		return
	}

	if b != '*' {
		return res, fmt.Errorf("Invalid first character: %q", b)
	} else {
		n, err := readLength(buf)
		if err != nil {
			return res, err
		}

		res = make([][]byte, n)

		for i := int64(0); i < n; i++ {
			res[i], err = readBulk(buf)
			if err != nil {
				return res, err
			}
		}

		b, err = buf.ReadByte()
		if err != nil && err != io.EOF {
			return res, err
		} else if err != io.EOF {
			return res, fmt.Errorf("Trailing garbage")
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
func (c *Client) Read() (interface{}, error) {
	return gedis.Read(*c.conn)
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
	return c.Write([]byte("+" + status + "\r\n"))
}

// Signature that command handler functions must have
type Handler func(c *Client, args []string) error

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
				fmt.Printf("Error while reading from client: %v\n", err)
			}
			return
		}

		data := in.([]interface{})
		cmd := strings.ToUpper(data[0].(string))
		args := make([]string, len(data)-1)

		for i, arg := range data[1:] {
			args[i] = fmt.Sprintf("%v", arg)
		}

		if fn := s.handlers[cmd]; fn != nil {
			err = fn(c, args)

			if err != nil {
				fmt.Printf("Unexpected error while processing connection: %v\n", err)
			}
		} else {
			c.Errorf("Unrecognized command '%s'", cmd)
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
