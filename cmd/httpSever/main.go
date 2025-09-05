package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"tcp.scratch.i/internal/response"
	"tcp.scratch.i/internal/server"
	request "tcp.scratch.i/internal/tests"
)

const port = 8080

func respond400() []byte {
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

func respond200() []byte {
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

func respond500() []byte {
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

func main() {
	s, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOk

		switch {
		case req.RequestLine.RequestTarget == "/yourproblem":
			body = respond400()
			status = response.StatusBadRequest

		case req.RequestLine.RequestTarget == "/myproblem":
			body = respond500()
			status = response.StatusInternalSeverError

		case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream"):
			requestTarget := req.RequestLine.RequestTarget

			res, err := http.Get("https://httpbin.org/" + requestTarget[len("/httpbin/"):])
			if err != nil {
				body = respond500()
				status = response.StatusInternalSeverError

				h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
				h.Replace("Content-Type", "text/html")

				w.WriteStatusLine(status)
				w.WriteHeaders(*h)
				w.WriteBody(body)
				return
			} else {

				w.WriteStatusLine(response.StatusOk)

				h.Delete("Content-Length")
				h.Set("transfer-encoding", "chunked")
				h.Replace("Content-Type", "text/plain")

				w.WriteHeaders(*h)

				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}

				w.WriteBody([]byte("0\r\n\r\n"))
			}
			return

		}

		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")

		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Server gracefully stopped")
}
