# gedis - a low-level interface in Go for Redis

`gedis` implements a very low-level interface to read and write using the [Redis](http://redis.io/) [protocol](http://redis.io/topics/protocol).

It also provides a simple client to talk to a Redis server and a generic Redis server, which allows you to implement your own server that understands the Redis protocol.

[![Build Status](https://travis-ci.org/inkel/gedis.png?branch=master)](https://travis-ci.org/inkel/gedis) `master` branch status at [Travis CI](https://travis-ci.org/)

## API documentation

gedis API documentation is available at the fabulous [GoDoc](http://godoc.org/) website, in the following locations:

* Parser: http://godoc.org/github.com/inkel/gedis
* Server: http://godoc.org/github.com/inkel/gedis/server
* Client: http://godoc.org/github.com/inkel/gedis/client

## Examples

You can find all the examples at https://github.com/inkel/gedis-examples

### Parser

In this example we'll create [`net.Conn`](http://golang.org/pkg/net/#Conn) to a Redis server and we'll send commands by using the parser function [`gedis.Write`](http://godoc.org/github.com/inkel/gedis#Write), and then read the server's response by using [`gedis.Read`](http://godoc.org/github.com/inkel/gedis#Read):

```go
package main

import (
	"flag"
	"fmt"
	"github.com/inkel/gedis"
	"net"
)

var server = flag.String("s", "localhost:6379", "Address of the Redis server")

func main() {
	c, err := net.Dial("tcp", *server)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	f := func(args ...interface{}) {
		cmd := args[0]

		fmt.Printf("> %s", cmd)
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
}
```

#### Output

```
$ go run ./gedis.go
> PING
< "PONG"

> SET "lorem" "ipsum"
< "OK"

> INCR "counter"
< 6

> HMSET "hash" "field1" "lorem" "field2" "ipsum"
< "OK"

> HGETALL "hash"
< []interface {}{"field1", "lorem", "field2", "ipsum"}
```

### Client

You can also use the [`gedis` client](http://godoc.org/github.com/inkel/gedis/client) package to create a [`Client`](http://godoc.org/github.com/inkel/gedis/client#Client) object that exposes almost the same API as using a `net.Conn`. In the future this client might add more features.

```go
package main

import (
	"flag"
	"fmt"
	"github.com/inkel/gedis/client"
)

var server = flag.String("s", "localhost:6379", "Address of the Redis server")

func main() {
	c, err := client.Dial("tcp", *server)
	if err != nil {
		panic(err)
		return
	}
	defer c.Close()

	f := func(args ...interface{}) {
		fmt.Printf("> %v\n", args)

		// Send to command to Redis server
		res, err := c.Send(args...)
		if err != nil {
			panic(err)
		} else {
			fmt.Printf("< %#v\n\n", res)
		}
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

#### Output

```
$ go run ./client.go
> [PING]
< "PONG"

> [SET lorem ipsum]
< "OK"

> [INCR counter]
< 7

> [HMSET hash field1 lorem field2 ipsum]
< "OK"

> [HGETALL hash]
< []interface {}{"field1", "lorem", "field2", "ipsum"}

> [MULTI]
< "OK"

> [GET counter]
< "QUEUED"

> [GET nonexisting]
< "QUEUED"

> [EXEC]
< []interface {}{"7", interface {}(nil)}
```

### Server

If you want to build a custom server that understands the Redis protocol, you can use the [`Server`](http://godoc.org/github.com/inkel/gedis/server#Server) type defined in the [`gedis` server](http://godoc.org/github.com/inkel/gedis/server) namespace.

The following example implements a server that only responds to the [`PING`](http://redis.io/commands/ping) command:

```go
package main

import (
	"flag"
	"fmt"
	"github.com/inkel/gedis/server"
	"os"
	"os/signal"
)

var listen = flag.String("l", ":26379", "Address to listen for connections")

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s, err := server.NewServer("tcp", *listen)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	pong := []byte("+PONG\r\n")
	earg := []byte("-ERR wrong number of arguments for 'ping' command\r\n")

	s.Handle("PING", func(c *server.Client, args [][]byte) error {
		if len(args) != 0 {
			c.Write(earg)
			return nil
		} else {
			c.Write(pong)
		}

		return nil
	})

	go s.Loop()

	// Wait for interrupt/kill
	<-c

	fmt.Println("Bye!")
}
```

#### Usage

```
$ go run ./server.go &
$ redis-cli -p 26379
redis 127.0.0.1:26379> ping
PONG
redis 127.0.0.1:26379> get inkel
(error) ERR Unrecognized command 'get'
redis 127.0.0.1:26379>
```

This generic server performs quite well, though not as fast as the standard C Redis server (which is kind of obvious):

```
$ redis-benchmark -q -t PING_MBULK -p 26379
PING_BULK: 36630.04 requests per second
```

## Build & test

In your `$GOPATH` do the following:

```
go get github.com/inkel/gedis
go get github.com/inkel/gedis/client
go get github.com/inkel/gedis/server
```

Then you can build it by executing:

```
go build github.com/inkel/gedis
go build github.com/inkel/gedis/client
go build github.com/inkel/gedis/server
```

Testing and benchmark:

```
go test github.com/inkel/gedis
go test github.com/inkel/gedis -bench=".*"

go test github.com/inkel/gedis/client
go test github.com/inkel/gedis/client -bench=".*"

go test github.com/inkel/gedis/server
go test github.com/inkel/gedis/server -bench=".*"
```

## References

* [Redis protocol](http://redis.io/topics/protocol)
* [How to write Go code](http://golang.org/doc/code.html)

## Why

I wanted to learn Go, so I decided to write a minimal Redis client.

Perhaps one day I might decide to do something else with it, but in the mean, the goals are merely academic.

Feel free to comment on the code and send patches if you like.

## License

```
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
```
