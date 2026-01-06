package response

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tsironi93/miniHttp/internal/headers"
)

const (
	Conn      = "Connection"
	ContLen   = "Content-Length"
	ContType  = "Content-Type"
	TransfEnc = "Transfer-Encoding"
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
	chunked    bool
}

func NewWriter() *Writer {
	h := GetDefaultHeaders()
	return &Writer{
		StatusCode: StatusOK,
		Headers:    h,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.Body.Write(p)
}

func (w *Writer) WriteString(s string) (int, error) {
	return w.Body.WriteString(s)
}

func (w *Writer) WriteChunkedBody(p []byte, out io.Writer) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	sizeLine := strconv.FormatInt(int64(len(p)), 16) + "\r\n"
	if n, err := io.WriteString(out, sizeLine); err != nil {
		return n, err
	}

	n, err := out.Write(p)
	if err != nil {
		return 0, nil
	}

	if _, err := io.WriteString(out, "\r\n"); err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone(out io.Writer) (int, error) {
	n, err := io.WriteString(out, "0\r\n\r\n")
	return n, err
}

func (w *Writer) WriteResponse(out io.Writer) error {
	if err := w.WriteStatusLine(out); err != nil {
		return err
	}

	if err := w.WriteHeaders(out); err != nil {
		return err
	}

	if _, err := w.WriteBody(out); err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders() headers.Headers {
	h := headers.NewHeaders()
	h[Conn] = "close"
	h[ContType] = "text/html"

	return h
}

func (w *Writer) WriteStatusLine(out io.Writer) error {
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

	if _, err := fmt.Fprintf(out, "HTTP/1.1 %d %s\r\n", w.StatusCode, text); err != nil {
		return err
	}

	w.state = stateStatusWritten
	return nil
}

func (w *Writer) WriteHeaders(out io.Writer) error {
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

	if _, err := io.WriteString(out, headerStr.String()); err != nil {
		return err
	}

	w.state = stateHeadersWritten
	return nil
}

func (w *Writer) WriteBody(out io.Writer) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("WriteBody called out of order")
	}

	n, err := out.Write(w.Body.Bytes())
	if err != nil {
		return 0, err
	}

	w.state = stateBodyWritten
	return n, nil
}
