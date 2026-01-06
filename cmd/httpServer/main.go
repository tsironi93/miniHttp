package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/tsironi93/miniHttp/internal/headers"
	"github.com/tsironi93/miniHttp/internal/request"
	"github.com/tsironi93/miniHttp/internal/response"
	"github.com/tsironi93/miniHttp/internal/server"
)

const (
	port          = 42069
	targetHTTPBin = "/httpbin"
	HTTPBinUrl    = "https://httpbin.org"
)

func loadHtml(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	return string(data)
}

func handleVideo(w *response.Writer) {
	videoPath := "./assets/vim.mp4"

	data, err := os.ReadFile(videoPath)
	if err != nil {
		log.Println(err)
		response.WriteBadRequestResponse(w.Out)
		return
	}

	h :=response.GetDefaultHeaders()
	h[response.ContType] = "video/mp4"
	w.WriteHeaders()

	w.Write(data)
	w.WriteResponse()
}

func htmlHandler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.StatusCode = response.StatusBadRequest
		w.WriteString(loadHtml("./internal/htmlTemplates/400.html"))
	case "/myproblem":
		w.StatusCode = response.StatusInternalServerError
		w.WriteString(loadHtml("./internal/htmlTemplates/500.html"))
	case "/video":
		handleVideo(w)
	default:
		w.StatusCode = response.StatusOK
		w.WriteString(loadHtml("./internal/htmlTemplates/200.html"))
	}
	w.WriteResponse()
}

func handleHTTPBinProxy(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, targetHTTPBin)
	if path == "" {
		path = "/"
	}

	targetURL := HTTPBinUrl + path

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
	w.Headers["Trailer"] = "X-Content-SHA256, X-Content-Length"
	w.Headers[response.ContType] = resp.Header.Get(response.ContType)

	if err := w.WriteStatusLine(); err != nil {
		return
	}

	if err := w.WriteHeaders(); err != nil {
		return
	}

	buf := make([]byte, 1024)
	hasher := sha256.New()
	totalLen := 0

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			hasher.Write(chunk)
			totalLen += n

			_, writeErr := w.WriteChunkedBody(chunk)
			if writeErr != nil {
				log.Println("Error writing chunk:", writeErr)
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
	w.WriteChunkedBodyDone()
	trailers := headers.NewHeaders()
	trailers["X-Content-SHA256"] = hex.EncodeToString(hasher.Sum(nil))
	trailers["X-Content-Length"] = strconv.Itoa(totalLen)
	w.WriteTrailers(trailers)
}

func mainHandler(w *response.Writer, req *request.Request) {
	if req.Headers[response.ContType] == "chunked" {
		w.Chunked = true
		handleHTTPBinProxy(w, req)
	} else {
		htmlHandler(w, req)
	}
}

func main() {
	server, err := server.Serve(port, mainHandler)
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
