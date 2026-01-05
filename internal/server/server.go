package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/tsironi93/miniHttp/internal/request"
	"github.com/tsironi93/miniHttp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  HandlerFunc
	closed   atomic.Bool
}

type HandleError struct {
	StatusCode response.StatusCode
	Msg        string
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

type HandlerFunc func(w *response.Writer, req *request.Request)

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		errWriter := response.NewWriter()
		errWriter.StatusCode = response.StatusBadRequest
		errWriter.Body.Reset()
		errWriter.Body.Write([]byte("Bad Request\n"))
		errWriter.WriteResponse(conn)
		return
	}

	rw := response.NewWriter()

	s.handler(rw, req)
	rw.WriteResponse(conn)
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
			if s.closed.Load() {
				return
			}
		}
		go s.handle(conn)
	}
}

func Serve(port int, handler HandlerFunc) (*Server, error) {

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	s := &Server{
		listener: ln,
		handler:  handler,
	}

	go s.listen()

	return s, nil
}
