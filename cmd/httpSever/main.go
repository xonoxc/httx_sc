package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"tcp.scratch.i/internal/server"
)

const port = 8080

func main() {
	s, err := server.Serve(port)
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
