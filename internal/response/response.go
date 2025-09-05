// Package response provides utilities for creating standardized API responses.
package response

import (
	"fmt"
	"io"
	"strconv"

	"tcp.scratch.i/internal/headers"
)

type StatusCode int

const (
	StatusOk                 StatusCode = 200
	StatusNotFound           StatusCode = 404
	StatusNotAuthorized      StatusCode = 401
	StatusInternalSeverError StatusCode = 500
	StatusCreated            StatusCode = 201
	StatusBadRequest         StatusCode = 400
)

type Response struct {
	Status StatusCode
}

// GetDefaultHeaders function set the default headers (until overwitten)
func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "closed")
	h.Set("Content-Type", "text/plain")

	return h
}

// Writer the part further is writer one where i have the custom writer
// to give users greater flexiblity of setting and playing with the headers
type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

// WriteStatusLine is the status line in this case isn't same as request line
// this is of the format : http-version http-status  status-text
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}

	switch statusCode {
	case StatusOk:
		statusLine = []byte("Http/1.1 200 OK\r\n")
	case StatusCreated:
		statusLine = []byte("Http/1.1 201 Created\r\n")
	case StatusNotAuthorized:
		statusLine = []byte("Http/1.1 401 Unauthorized\r\n")
	case StatusInternalSeverError:
		statusLine = []byte("Http/1.1 500 Internal Sever Error\r\n")
	case StatusNotFound:
		statusLine = []byte("Http/1.1 404 Not Found\r\n")
	case StatusBadRequest:
		statusLine = []byte("Http/1.1 400 Bad Request\r\n")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	h.Map(func(k, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	})
	b = fmt.Append(b, "\r\n")

	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.writer.Write(p)
}

// status Responses
func Respond400() []byte {
	return []byte(`
		<html>
		  <head>
			<title>200 OK</title>
		  </head>
		  <body>
			<h1>Success!</h1>
			<p>Your request was an absolute banger.</p>
		  </body>
		</html>
	`)
}

func Respond200() []byte {
	return []byte(`
		 <html>
		  <head>
			<title>500 Internal Server Error</title>
		  </head>
		  <body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		  </body>
		</html>
	`)
}

func Respond500() []byte {
	return []byte(`
		 <html>
		  <head>
			<title>500 Internal Server Error</title>
		  </head>
		  <body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		  </body>
		</html>
	`)
}
