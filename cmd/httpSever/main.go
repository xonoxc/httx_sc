package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"tcp.scratch.i/internal/headers"
	"tcp.scratch.i/internal/response"
	"tcp.scratch.i/internal/server"
	request "tcp.scratch.i/internal/tests"
)

const port = 8080

func toStr(payload []byte) string {
	out := ""
	for _, b := range payload {
		out += fmt.Sprintf("%x", b)
	}
	return out
}

func main() {
	s, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := response.Respond200()
		status := response.StatusOk

		switch {
		case req.RequestLine.RequestTarget == "/yourproblem":
			body = response.Respond400()
			status = response.StatusBadRequest

		case req.RequestLine.RequestTarget == "/myproblem":
			body = response.Respond500()
			status = response.StatusInternalSeverError

		case req.RequestLine.RequestTarget == "/video":
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOk)
			w.WriteHeaders(*h)
			w.WriteBody(f)

		case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream"):
			requestTarget := req.RequestLine.RequestTarget

			res, err := http.Get("https://httpbin.org/" + requestTarget[len("/httpbin/"):])
			if err != nil {
				body = response.Respond500()
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

				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")

				w.WriteHeaders(*h)

				fullbody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					fullbody = append(fullbody, data[:n]...)
					w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))

				trailers := headers.NewHeaders()

				out := sha256.Sum256(fullbody)

				trailers.Set("X-Content-SHA256", toStr(out[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullbody)))

				w.WriteHeaders(*trailers)
				w.WriteBody([]byte("\r\n"))

				return
			}

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
