package main

import (
	"fmt"
	"github.com/tsironi93/miniHttp/internal/request"
	"net"
)

func printRequest(r *request.Request) {

	fmt.Println("Request Line:")
	fmt.Println("- Method:", r.RequestLine.Method)
	fmt.Println("- Target:", r.RequestLine.RequestTarget)
	fmt.Println("- Version:", r.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for k, v := range r.Headers {
		fmt.Println("-", k+":", v)
	}
	fmt.Println("Body:")
	fmt.Println(string(r.Body))
}

func main() {

	listener, er := net.Listen("tcp", "127.0.0.1:42069")
	if er != nil {
		panic(er)
	}
	defer listener.Close()

	for {
		fd, er := listener.Accept()
		if er != nil {
			panic(er)
		}

		req, err := request.RequestFromReader(fd)
		if err != nil {
			fmt.Println("parse error:", err)
		}

		printRequest(req)

		fd.Close()
	}
}
