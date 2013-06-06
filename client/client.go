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
gedis client - Redis client written in Go

This package allows to create clients that can talk to servers by using the Redis protocol

Redis protocol: http://redis.io/topics/protocol

Example

    package main

    import (
    	"fmt"
    	"github.com/inkel/gedis/client"
    )

    func main() {
    	c, err := client.Dial("tcp", "localhost:6379")
    	if err != nil {
    		panic(err)
    	}
    	defer c.Close()

    	f := func(args ...string) {
    		fmt.Printf("> %s", args[0])
    		for _, arg := range args[1:] {
    			fmt.Printf(" %q", arg)
    		}
    		fmt.Println()

    		// Send to command to Redis server
    		res, err := c.Send(args...)
    		if err != nil {
    			panic(err)
    		}

    		fmt.Printf("< %#v\n\n", res)
    	}

    	f("PING")

    	f("SET", "lorem", "ipsum")

    	f("INCR", "counter")

    	f("HMSET", "hash", "field1", "lorem", "field2", "ipsum")

    	f("HGETALL", "hash")
    }

*/
package client

import (
	"github.com/inkel/gedis"
	"net"
)

// A wrapper to net.Conn that handles writing/reading to a Redis
// server
type Client struct {
	conn net.Conn
}

// Connect to a Redis server on address, using the named network
//
// See documentation for net.Dial for more information on named
// networks.
func Dial(network, address string) (c Client, err error) {
	c.conn, err = net.Dial(network, address)
	return
}

// Close the connection to the Redis server
func (c *Client) Close() error {
	return c.conn.Close()
}

// Send a command to the Redis server and receive its reply
func (c *Client) Send(args ...interface{}) (interface{}, error) {
	_, err := gedis.Write(c.conn, args...)
	if err != nil {
		return nil, err
	}
	return c.Read()
}

// Reads from the client
//
// This is useful for cases like a monitor
func (c *Client) Read() (interface{}, error) {
	return gedis.Read(c.conn)
}
