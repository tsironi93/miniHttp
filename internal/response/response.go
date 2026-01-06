package response

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/tsironi93/miniHttp/internal/headers"
)

const (
	HTML      = "text/html"
	Conn      = "Connection"
	CRLF      = "\r\n"
	ContLen   = "Content-Length"
	ContType  = "Content-Type"
	TransfEnc = "Transfer-Encoding"
)

type ContentType int

const (
	Plain ContentType = iota
	Html
	Image
	Video
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	stateInit writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
)

type Writer struct {
	StatusCode StatusCode
	Headers    map[string]string
	Body       bytes.Buffer
	state      writerState
	Chunked    bool
	Out       net.Conn
}

func NewWriter(out net.Conn) *Writer {
	h := GetDefaultHeaders()
	return &Writer{
		StatusCode: StatusOK,
		Headers:    h,
		Chunked:    false,
		Out:       out,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.Body.Write(p)
}

func (w *Writer) WriteString(s string) (int, error) {
	return w.Body.WriteString(s)
}

func (w *Writer) WriteResponse() error {
	if err := w.WriteStatusLine(); err != nil {
		return err
	}

	if err := w.WriteHeaders(); err != nil {
		return err
	}

	if _, err := w.WriteBody(); err != nil {
		return err
	}

	return nil
}

func WriteBadRequestResponse(out net.Conn) {
	errWriter := NewWriter(out)
	errWriter.StatusCode = StatusBadRequest
	errWriter.Body.Reset()
	errWriter.Body.Write([]byte("Bad Request\n"))
	errWriter.WriteResponse()
}

func GetDefaultHeaders() headers.Headers {
	h := headers.NewHeaders()
	h[Conn] = "close"
	h[ContType] = HTML

	return h
}

func (w *Writer) WriteStatusLine() error {
	if w.state != stateInit {
		return fmt.Errorf("WriteStatusLine called out of order")
	}

	statusText := map[StatusCode]string{
		StatusOK:                  "OK",
		StatusBadRequest:          "Bad Request",
		StatusInternalServerError: "Internal Server Error",
	}

	text, ok := statusText[w.StatusCode]
	if !ok {
		text = "Unknown"
	}

	if _, err := fmt.Fprintf(w.Out, "HTTP/1.1 %d %s"+CRLF, w.StatusCode, text); err != nil {
		return err
	}

	w.state = stateStatusWritten
	return nil
}

func (w *Writer) WriteHeaders() error {
	if w.state != stateStatusWritten {
		return fmt.Errorf("WriteHeaders called out of order")
	}

	if _, ok := w.Headers[ContLen]; !ok {
		w.Headers[ContLen] = strconv.Itoa(len(w.Body.Bytes()))
	}

	var headerStr strings.Builder
	for k, v := range w.Headers {
		headerStr.WriteString(k)
		headerStr.WriteString(": ")
		headerStr.WriteString(v)
		headerStr.WriteString("\r\n")
	}
	headerStr.WriteString("\r\n")

	if _, err := io.WriteString(w.Out, headerStr.String()); err != nil {
		return err
	}

	w.state = stateHeadersWritten
	return nil
}

func (w *Writer) WriteBody() (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("WriteBody called out of order")
	}

	n, err := w.Out.Write(w.Body.Bytes())
	if err != nil {
		return 0, err
	}

	w.state = stateBodyWritten
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	sizeLine := strconv.FormatInt(int64(len(p)), 16) + CRLF
	if n, err := io.WriteString(w.Out, sizeLine); err != nil {
		return n, err
	}

	n, err := w.Out.Write(p)
	if err != nil {
		return 0, nil
	}

	if _, err := io.WriteString(w.Out, CRLF); err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := io.WriteString(w.Out, "0"+CRLF)
	w.state = stateBodyWritten
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != stateBodyWritten {
		return fmt.Errorf("WriteTrailers called out of order")
	}

	var b strings.Builder
	for k, v := range h {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v)
		b.WriteString(CRLF)
	}
	b.WriteString(CRLF)

	_, err := io.WriteString(w.Out, b.String())
	return err
}
