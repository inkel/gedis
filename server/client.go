package server

import (
	"fmt"
	"github.com/inkel/gedis"
	"net"
)

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
func (c *Client) Read() ([][]byte, error) {
	return Read(*c.conn)
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
	return c.Write(gedis.WriteStatus(status))
}
