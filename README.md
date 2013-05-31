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
