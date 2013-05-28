# gedis - a Redis client written in Go

I wanted to learn Go and network programming, so I decided to write a minimal Redis client.

**This is far from being production ready.** Perhaps one day I might decide to do something else with it, but in the mean, the goals are merely academic.

Feel free to comment on the code and send patches if you like.

## References

* [Redis protocol](http://redis.io/topics/protocol)

## TODO

* Documentation
* Tests
  * `\r\n` in a reply
  * Null elements in Multi-Bulk replies
* Network
* Socket
* Pipeline
* Improve Bulk/Multi-Bulk `nil`