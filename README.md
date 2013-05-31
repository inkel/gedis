# gedis - a Redis client written in Go

I wanted to learn Go and network programming, so I decided to write a minimal Redis client.

**This is far from being production ready.** Perhaps one day I might decide to do something else with it, but in the mean, the goals are merely academic.

Feel free to comment on the code and send patches if you like.

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

## API

Gedis currently provides API functions for writing and reading the [Redis protocol](http://redis.io/topics/protocol).

### Writing

* `WriteBulk(bulk string) []byte`: writes the `bulk` string into a `[]byte`
* `WriteMultiBulk(cmd string, args …string) []byte`: writes a multi-bulk sequence into a `[]byte`. **This method will probably change it's name any time soon.**

### Reading

* `ReadBulk(r io.Reader) (bytes []byte, err error)`: reads the next bulk from `r`, an [`io.Reader`](http://golang.org/pkg/io/#Reader) and returns a `[]byte`, or an error. Note that this function assumes that the bulk identification character `$` was already detected. See the implementation of `Parse`. **This function should probably be private.**
* `Parse(r io.Reader) (ret interface{}, err error)`: parses a full response from Redis. It will return each type as expected:
  * status and bulk replies will be returned as strings
  * integer replies will be returned as a signed integer
  * null replies will return a `nil`
  * multi-bulk replies will return a `[]interface{}` where each element will be properly represented as above.

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
