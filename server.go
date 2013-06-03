package gedis

import (
	"fmt"
	"net"
	"strings"
)

type Handler func(c net.Conn, args []string) error

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

func (s *Server) process(c net.Conn) {
	in, err := Read(c)
	if err != nil {
		fmt.Printf("Error while reading from client: %v\n", err)
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
		s.Errorf(c, "Unrecognized command '%s'", cmd)
	}
}

func (s *Server) Loop() {
	for {
		client, err := s.ln.Accept()
		if err != nil {
			fmt.Printf("Error while accepting a connection: %v\n", err)
			continue
		}
		go s.process(client)
	}
}

func (s *Server) Error(c net.Conn, err error) {
	c.Write(writeError(err))
}

func (s *Server) Errorf(c net.Conn, format string, args ...interface{}) {
	s.Error(c, fmt.Errorf(format, args...))
}

func (s *Server) Status(c net.Conn, status string) {
	c.Write([]byte("+" + status + "\r\n"))
}

func (s *Server) Bulk(c net.Conn, bulk string) {
	c.Write(writeBulk(bulk))
}

func (s *Server) Nil(c net.Conn) {
	c.Write([]byte("$-1\r\n"))
}

func (s *Server) Int(c net.Conn, n int64) {
	c.Write(writeInt(n))
}
