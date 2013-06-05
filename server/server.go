package server

import (
	"fmt"
	"github.com/inkel/gedis"
	"io"
	"net"
	"strings"
)

// Holds pointers to the current Server and client net.Conn
type Client struct {
	server *Server
	conn   *net.Conn
}

func (c *Client) Close() {
	conn := *c.conn
	conn.Close()
}

// Read from the client, parsing the input with the Redis protocol
func (c *Client) Read() (interface{}, error) {
	return gedis.Read(*c.conn)
}

//
func (c *Client) Write(bytes []byte) (int, error) {
	conn := *c.conn
	return conn.Write(bytes)
}

// Sends an error to the client
func (c *Client) Error(err error) (int, error) {
	return c.Write(gedis.WriteError(err))
}

// Sends a string formatted as an error to the client
func (c *Client) Errorf(format string, args ...interface{}) (int, error) {
	return c.Error(fmt.Errorf(format, args...))
}

func (c *Client) Status(status string) (int, error) {
	return c.Write([]byte("+" + status + "\r\n"))
}

type Handler func(c *Client, args []string) error

type Server struct {
	ln       net.Listener
	handlers map[string]Handler
}

func NewServer(network, address string) (s Server, err error) {
	s.handlers = make(map[string]Handler)
	s.ln, err = net.Listen(network, address)
	return
}

func (s *Server) Close() error {
	return s.ln.Close()
}

func (s *Server) Handle(cmd string, handler Handler) {
	cmd = strings.ToUpper(cmd)
	s.handlers[cmd] = handler
}

func (s *Server) process(c *Client) {
	defer c.Close()

	for {
		in, err := c.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error while reading from client: %v\n", err)
			}
			return
		}

		data := in.([]interface{})
		cmd := strings.ToUpper(data[0].(string))
		args := make([]string, len(data)-1)

		for i, arg := range data[1:] {
			args[i] = fmt.Sprintf("%v", arg)
		}

		if fn := s.handlers[cmd]; fn != nil {
			err = fn(c, args)

			if err != nil {
				fmt.Printf("Unexpected error while processing connection: %v\n", err)
			}
		} else {
			c.Errorf("Unrecognized command '%s'", cmd)
		}
	}
}

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
