package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/tsironi93/miniHttp/internal/request"
	"github.com/tsironi93/miniHttp/internal/response"
)

type Server struct {
	listener net.Listener
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

type HandlerFunc func(w io.Writer, req *request.Request) *HandleError

func (he HandleError) ErrorWriter(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("nil writer")
	}

	if err := response.WriteStatusLine(w, he.StatusCode); err != nil {
		return err
	}

	body := he.Msg + "\n"
	h := response.GetDefaultHeaders(len(he.Msg))
	if err := response.WriteHeaders(w, h); err != nil {
		return err
	}

	_, err := io.WriteString(w, body)
	return err
}

func (s *Server) handle(conn net.Conn, handler HandlerFunc) {
	defer conn.Close()

	r, err := request.RequestFromReader(conn)
	if err != nil {
		(&HandleError{
			StatusCode: response.SC400,
			Msg:        "bad request",
		}).ErrorWriter(conn)
		return
	}

	var buf bytes.Buffer
	if herr := handler(&buf, r); herr != nil {
		herr.ErrorWriter(conn)
		return
	}

	body := buf.Bytes()
	response.WriteResponse(conn, len(body), response.SC400)
}

func (s *Server) listen(handler HandlerFunc) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
			if s.closed.Load() {
				return
			}
		}
		go s.handle(conn, handler)
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
	}

	go s.listen(handler)

	return s, nil
}
