package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/tsironi93/miniHttp/internal/headers"
)

const ContLen = "Content-Length"
const Conn = "Connection"
const ContType = "Content-Type"

type StatusCode int

const (
	SC200 = iota
	SC400
	SC500
)

// func WriteErrorResponse(w io.Writer, code StatusCode, msg string) {
// 	body := msg + "\n"
//
// 	h := GetDefaultHeaders(len(body))
//
// 	WriteStatusLine(w, code)
// 	WriteHeaders(w, h)
// 	io.WriteString(w, body)
// }
//
// func WriteError(w io.Writer, err error) {
// 	var errHandler *HandleError
//
// 	if errors.As(err, &errHandler) {
// 		WriteErrorResponse(w, errHandler.Code, errHandler.Msg)
// 		return
// 	}
// }

func WriteResponse(w io.Writer, bodyLen int, status StatusCode) error {
	if w == nil {
		return fmt.Errorf("nil writer")
	}

	WriteStatusLine(w, status)
	h := GetDefaultHeaders(bodyLen)
	err := WriteHeaders(w, h)
	return err
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	s := fmt.Sprintf(
		"%s: %s\r\n%s: %s\r\n%s: %s\r\n\r\n",
		ContLen, h[ContLen],
		Conn, h[Conn],
		ContType, h[ContType],
	)

	_, err := io.WriteString(w, s)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h[ContLen] = strconv.Itoa(contentLen)
	h[Conn] = "close"
	h[ContType] = "text/plain"

	return h
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	if _, err := io.WriteString(w, "HTTP/1.1 "); err != nil {
		return err
	}

	switch statusCode {
	case SC200:
		_, err := io.WriteString(w, "200 OK\r\n")
		return err
	case SC400:
		_, err := io.WriteString(w, "400 Bad Request\r\n")
		return err
	case SC500:
		_, err := io.WriteString(w, "500 Internal Server Error\r\n")
		return err
	default:
		return nil
	}
}

// func (e *HandleError) Error() string {
// 	return e.Msg
// }
//
// func BadRequest(msg string) *HandleError {
// 	return &HandleError{
// 		Code: SC400,
// 		Msg:  msg,
// 	}
// }
//
// func InternalError(msg string) *HandleError {
// 	return &HandleError{
// 		Code: SC500,
// 		Msg:  msg,
// 	}
// }
