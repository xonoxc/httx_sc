// Package server provides a simple HTTP server and it corresponding methods
package server

import (
	"fmt"
	"io"
	"net"

	"tcp.scratch.i/internal/response"
	request "tcp.scratch.i/internal/tests"
)

type Server struct {
	closed  bool
	handler Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w response.Writer, req *request.Request)

func runRequest(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)

	respWriter := response.NewWriter(conn)
	if err != nil {
		respWriter.WriteStatusLine(response.StatusBadRequest)
		return
	}

	s.handler(*respWriter, req)
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

			go runRequest(s, conn)
		}
	}()
}

func Serve(port uint16, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:  false,
		handler: handler,
	}
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
