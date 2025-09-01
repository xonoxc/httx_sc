// Package server provides a simple HTTP server and it corresponding methods
package server

import (
	"fmt"
	"io"
	"net"
)

type Server struct {
	closed bool
}

func runRequest(conn io.ReadWriteCloser) {
	out := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!")
	conn.Write(out)
	conn.Close()
}

func runServer(s *Server, listener net.Listener) {
	go func() {
		for {
			conn, err := listener.Accept()
			if s.closed {
				return
			}

			if err != nil {
				return
			}

			go runRequest(conn)
		}
	}()
}

func Serve(port uint16) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	defer ln.Close()

	server := &Server{closed: false}
	go runServer(server, ln)

	return server, err
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}

func (s *Server) Listen() {
}

func (s *Server) Handle(conn net.Conn) {
}
