package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tsironi93/miniHttp/internal/request"
	"github.com/tsironi93/miniHttp/internal/response"
	"github.com/tsironi93/miniHttp/internal/server"
)

const port = 42069

func testHandler(w io.Writer, req *request.Request) *server.HandleError {

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandleError{
			StatusCode: response.SC400,
			Msg:        "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandleError{
			StatusCode: response.SC500,
			Msg:        "Woopsie, my bad\n",
		}
	default:
		return &server.HandleError{
			StatusCode: response.SC200,
			Msg:        "All good, frfr\n",
		}
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
