# gedis - a low-level interface for Redis written in Go

`gedis` implements a very low-level interface to read and write using the Redis protocol. It also provides a simple client to talk to a Redis server.

**This is far from being production ready.**

## API

Gedis currently provides two API functions for writing and reading the [Redis protocol](http://redis.io/topics/protocol). It also defines two simple interfaces: `Reader` and `Writer`

### Writing

Gedis defines the following `Writer` interface:

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

It is possible to send Redis commands to any object that implements that interface, i.e. [`net.Conn`](http://golang.org/pkg/net/#Conn), by using the following function:

```go
Write(w Writer, args ...string) (n int, err error)
```

### Reading

Gedis defines the following `Reader` interface:

```go
type Reader interface {
	Read(b []byte) (n int, err error)
}
```

It is possible to read Redis replies from any object that implements that interface, i.e. [`net.Conn`](http://golang.org/pkg/net/#Conn), by using the following function:

```go
Read(r Reader) (reply interface{}, err error)
```

### Example

```go
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

	f := func(cmd string, args ...string) {
		fmt.Printf("> %s", cmd)
		for _, arg := range args {
			fmt.Printf(" %q", arg)
		}
		fmt.Println()

		// Send to command to Redis server
		_, err := gedis.Write(c, cmd, args...)
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

```

## Client

To avoid you the hassle of having to pass the connction parameter in every call, `gedis` defines the following `Client` object:

```go
type Client struct {}

func Dial(network, address string) (c Client, err error)

func (c *Client) Close() error

func (c *Client) Send(cmd string, args ...string) (interface{}, error)
```

### Example

```go
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

	f := func(cmd string, args ...string) {
		fmt.Printf("> %s", cmd)
		for _, arg := range args {
			fmt.Printf(" %q", arg)
		}
		fmt.Println()

		// Send to command to Redis server
		res, err := c.Send(cmd, args...)
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
```

## Build & test

In your `$GOPATH` do the following:

```
go get github.com/inkel/gedis
```

Then you can build it by executing:

```
go build github.com/inkel/gedis
```

Testing and benchmark:

```
go test github.com/inkel/gedis
go test github.com/inkel/gedis -bench=".*"
```

Note that running the benchmarks **will** run the tests beforehand.


## References

* [Redis protocol](http://redis.io/topics/protocol)
* [How to write Go code](http://golang.org/doc/code.html)

## TODO

* Documentation
* Tests
  * ~~`\r\n` in a reply~~
  * ~~Null elements in Multi-Bulk replies~~
  * ~~Multi-Bulk inside Multi-Bulk replies~~
* Network
* Socket
* Pipeline
* ~~Improve Bulk/Multi-Bulk `nil`~~

## Why

I wanted to learn Go, so I decided to write a minimal Redis client.

Perhaps one day I might decide to do something else with it, but in the mean, the goals are merely academic.

Feel free to comment on the code and send patches if you like.
