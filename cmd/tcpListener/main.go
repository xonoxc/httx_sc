package main

import (
	"fmt"
	"log"
	"net"

	request "tcp.scratch.i/internal/tests"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error while spinning up the server:", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error parsing request", err)
		}

		fmt.Println("RequestLine:")
		fmt.Println("- Method:", req.RequestLine.Method)
		fmt.Println("- Target:", req.RequestLine.RequestTarget)
		fmt.Println("- Version:", req.RequestLine.HTTPVersion)

		fmt.Println("Headers:")
		req.Headers.Map(func(k, v string) {
			fmt.Println("key:", k, "value:", v)
		})

		conn.Close()
	}
}
