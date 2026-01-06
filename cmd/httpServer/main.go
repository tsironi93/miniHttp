package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tsironi93/miniHttp/internal/request"
	"github.com/tsironi93/miniHttp/internal/response"
	"github.com/tsironi93/miniHttp/internal/server"
)

const port = 42069

func loadHtml(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	return string(data)
}

func handleHTTPBinProxy(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	if path == "" {
		path = "/"
	}

	targetURL := "https://httpbin.org" + path

	resp, err := http.Get(targetURL)
	if err != nil {
		w.StatusCode = response.StatusInternalServerError
		w.WriteString("Upstream error\n")
		return
	}
	defer resp.Body.Close()

	w.StatusCode = response.StatusCode(resp.StatusCode)
	delete(w.Headers, response.ContLen)
	w.Headers[response.TransfEnc] = "chunked"
	w.Headers[response.ContType] = resp.Header.Get(response.ContType)

	if err := w.WriteStatusLine(w); err != nil {
		return
	}

	if err := w.WriteHeaders(w); err != nil {
		return
	}

	buf := make([]byte, 32)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Println(string(buf))
			_, writeErr := w.WriteChunkedBody(buf[:n], w)
			if writeErr != nil {
				return
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
	}
	w.WriteChunkedBodyDone(w)
}

func htmlHandler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.StatusCode = response.StatusBadRequest
		w.WriteString(loadHtml("./internal/htmlTemplates/400.html"))
	case "/myproblem":
		w.StatusCode = response.StatusInternalServerError
		w.WriteString(loadHtml("./internal/htmlTemplates/500.html"))
	default:
		w.StatusCode = response.StatusOK
		w.WriteString(loadHtml("./internal/htmlTemplates/200.html"))
	}
}

func main() {
	server, err := server.Serve(port, handleHTTPBinProxy)
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
