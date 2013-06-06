/*
Copyright 2013 Leandro López (inkel)

This package implements a very low-level interface to read and
write using the Redis protocol. It also provides a simple client to
talk to a Redis server.

THIS IS FAR FROM BEING PRODUCTION READY.

API

Gedis currently provides two API functions for writing and reading
the [Redis protocol](http://redis.io/topics/protocol). It also
defines two simple interfaces: `Reader` and `Writer`

Writing commands

Gedis defines the following `Writer` interface:

 type Writer interface {
         Write(p []byte) (n int, err error)
 }

It is possible to send Redis commands to any object that implements
that interface, i.e. [`net.Conn`](http://golang.org/pkg/net/#Conn),
by using the following function:

 Write(w Writer, args ...string) (n int, err error)

Reading

Gedis defines the following `Reader` interface:

 type Reader interface {
         Read(b []byte) (n int, err error)
 }

It is possible to read Redis replies from any object that implements
that interface, i.e. net.Conn, by using the following function:

 Read(r Reader) (reply interface{}, err error)

API usage example

    package main

    import (
    	"fmt"
    	"github.com/inkel/gedis"
    	"net"
    )

    func main() {
    	c, err := net.Dial("tcp", "localhost:6379")
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
    		_, err := gedis.Write(c, args...)
    		if err != nil {
    			panic(err)
    		}

    		// Read the reply from the server
    		res, err := gedis.Read(c)
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

    	f("MULTI")
    	f("GET", "counter")
    	f("GET", "nonexisting")
    	f("EXEC")
    }


Client

To avoid you the hassle of having to pass the connction parameter
in every call, `gedis` defines the following `Client` object:

 type Client struct {}
 func Dial(network, address string) (c Client, err error)
 func (c *Client) Close() error
 func (c *Client) Send(cmd string, args ...string) (interface{}, error)

Client usage example

    package main

    import (
    	"fmt"
    	"github.com/inkel/gedis"
    )

    func main() {
    	c, err := gedis.Dial("tcp", "localhost:6379")
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

    	f("MULTI")
    	f("GET", "counter")
    	f("GET", "nonexisting")
    	f("EXEC")
    }


Why

I wanted to learn Go, so I decided to write a minimal Redis
client. Perhaps one day I might decide to do something else with it,
but in the mean, the goals are merely academic. Feel free to comment
on the code and send patches if you like.

Redis protocol

Redis uses a very simple text protocol, which is binary safe.

    *<num args> CR LF
    $<num bytes arg1> CR LF
    <arg data> CR LF
    ...
    $<num bytes argn> CR LF
    <arg data>
*/
package gedis

// Struct to hold parsing errors
type ParseError struct {
	err string
}

func (pe *ParseError) Error() string {
	return pe.err
}

func NewParseError(err string) *ParseError {
	return &ParseError{err}
}
