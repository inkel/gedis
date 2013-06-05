package server

import (
	"bytes"
	"fmt"
	"github.com/inkel/gedis"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

type Reader interface {
	Read([]byte) (int, error)
}

func readLength(buf *bytes.Buffer) (n int64, err error) {
	sn, err := buf.ReadString('\r')

	if err != nil {
		return -1, err
	}

	b, err := buf.ReadByte()
	if err != nil {
		return -1, err
	} else if b != '\n' {
		return -1, fmt.Errorf("Invalid EOL: %q", []byte{'\r', b})
	}

	if n < 0 {
		return -1, fmt.Errorf("Negative length: %d", n)
	}

	return strconv.ParseInt(sn[:len(sn)-1], 10, 64)
}

func readBulk(buf *bytes.Buffer) (bs []byte, err error) {
	var b byte

	b, err = buf.ReadByte()
	if err != nil {
		return bs, err
	} else if b != '$' {
		return bs, fmt.Errorf("Invalid first character: %q", b)
	}

	n, err := readLength(buf)
	if err != nil {
		return bs, err
	}

	bs = make([]byte, n)

	_, err = buf.Read(bs)
	if err != nil {
		return bs, err
	}

	crlf := make([]byte, 2)

	if _, err = buf.Read(crlf); err != nil {
		return bs, err
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return bs, fmt.Errorf("Invalid EOL: %q", crlf)
	}

	return
}

func Read(r Reader) (res [][]byte, err error) {
	var b byte

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)

	b, err = buf.ReadByte()
	if err != nil {
		return
	}

	if b != '*' {
		return res, fmt.Errorf("Invalid first character: %q", b)
	} else {
		n, err := readLength(buf)
		if err != nil {
			return res, err
		}

		res = make([][]byte, n)

		for i := int64(0); i < n; i++ {
			res[i], err = readBulk(buf)
			if err != nil {
				return res, err
			}
		}

		b, err = buf.ReadByte()
		if err != nil && err != io.EOF {
			return res, err
		} else if err != io.EOF {
			return res, fmt.Errorf("Trailing garbage")
		}
	}

	return res, err
}

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
