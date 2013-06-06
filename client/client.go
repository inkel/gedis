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
