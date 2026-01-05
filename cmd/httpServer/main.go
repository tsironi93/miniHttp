package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tsironi93/miniHttp/internal/request"
	"github.com/tsironi93/miniHttp/internal/response"
	"github.com/tsironi93/miniHttp/internal/server"
)

const port = 42069

func testHandler(w *response.Writer, req *request.Request) {

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.StatusCode = response.StatusBadRequest
		w.Write([]byte("Your problem is not my problem\n"))
	case "/myproblem":
		w.StatusCode = response.StatusInternalServerError
		w.Write([]byte("Woopsie, my bad\n"))
	default:
		w.StatusCode = response.StatusOK
		w.Write([]byte("All good, frfr\n"))
	}
}

func main() {
	server, err := server.Serve(port, testHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
