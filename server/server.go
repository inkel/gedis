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

    	s.Handle("PING", func(c *gedis.Client, args [][]byte) error {
    		if len(args) != 0 {
    			c.Write(earg)
    			return nil
    		} else {
    			c.Write(pong)
    		}

    		return nil
    	})

    	go s.Loop()

    	<-c

    	fmt.Println("Bye!")
    }
*/
package server

import (
	"fmt"
	"io"
	"net"
	"strings"
)

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
