package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/tsironi93/miniHttp/internal/headers"
)

const bufferSize int = 1024
const crlf = "\r\n"
const contLen = "content-length"

type parseState int

const (
	INITIALIZED = iota
	PARSING_HEADERS
	PARSING_BODY
	DONE
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       State
}

type State struct {
	parseState parseState
	dataRead   uint64
	dataParced uint64
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func isKeywordCapitalized(key string) bool {
	for _, c := range key {
		if unicode.IsLetter(c) && !unicode.IsUpper(c) {
			return false
		}
	}
	return true
}

func parceBody(r *Request, data []byte) int {
	bytesRead := 0
	cLen := 0
	remaining := 0
	v, ok := r.Headers.Get(contLen)

	if ok {
		cLen, _ = strconv.Atoi(v)
		remaining = cLen - len(r.Body)
		if remaining <= 0 {
			r.State.parseState = DONE
			return bytesRead
		}
	} else if !ok {
		r.State.parseState = DONE
		return bytesRead
	}

	toCopy := data
	if len(toCopy) > remaining {
		toCopy = data[:remaining]
	}

	r.Body = append(r.Body, toCopy...)
	bytesRead += len(toCopy)

	if len(r.Body) == cLen {
		r.State.parseState = DONE
	}

	return bytesRead
}

func (r *Request) parse(data []byte) (int, error) {

	if r.State.parseState == DONE {
		return -1, fmt.Errorf("error: trying to read data in DONE state")
	}

	if r.State.parseState != INITIALIZED &&
		r.State.parseState != PARSING_HEADERS &&
		r.State.parseState != PARSING_BODY {
		return -2, fmt.Errorf("error: unknown state")
	}

	totalBytes := 0

	if r.State.parseState == INITIALIZED {
		reqBytes, err := parseRequestLine(string(data), r)
		if err != nil {
			return -3, err
		}

		if reqBytes == 0 {
			return 0, nil
		}

		totalBytes += reqBytes
		r.State.parseState = PARSING_HEADERS
		data = data[reqBytes:]
	}

	for r.State.parseState == PARSING_HEADERS {

		headBytes, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.State.parseState = PARSING_BODY
		}

		if headBytes == 0 && !done {
			break
		}

		totalBytes += headBytes
		data = data[headBytes:]
	}

	if r.State.parseState == PARSING_BODY {
		totalBytes += parceBody(r, data)
	}
	return totalBytes, nil
}

func parseRequestLine(line string, r *Request) (int, error) {

	idx := strings.Index(line, crlf)
	if idx == -1 {
		return 0, nil
	}

	requestLine := line[:idx]
	parts := strings.Fields(requestLine)

	if len(parts) != 3 {
		return 0, fmt.Errorf("Invalid request: %q", requestLine)
	}

	method, target, version := parts[0], parts[1], parts[2]

	if !isKeywordCapitalized(method) || !isKeywordCapitalized(version) {
		return 0, fmt.Errorf("version or method are not capitalized: %s", line)
	}

	if method != "GET" && method != "POST" && method != "HEAD" {
		return 0, fmt.Errorf("wrong method or non existand: %s", method)
	}

	if !strings.HasPrefix(target, "/") {
		return 0, fmt.Errorf("Invalid target: %s", target)
	}

	if version != "HTTP/1.1" {
		return 0, fmt.Errorf("Wrong HTTP version requested: %s", version)
	}

	r.RequestLine.Method = method
	r.RequestLine.RequestTarget = target
	r.RequestLine.HttpVersion = "1.1"

	return idx + 2, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, bufferSize)
	readToIndex := 0

	r := &Request{
		State: State{
			parseState: INITIALIZED,
			dataRead:   0,
			dataParced: 0,
		},
		Headers: headers.NewHeaders(),
	}

	for r.State.parseState != DONE {

		if readToIndex == len(buf) {
			newBuf := make([]byte, 2*len(buf))
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			return nil, err
		}

		readToIndex += n
		r.State.dataRead += uint64(n)
		consumed, parseErr := r.parse(buf[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}

		if consumed > 0 {
			copy(buf, buf[consumed:readToIndex])
			readToIndex -= consumed
			r.State.dataParced += uint64(consumed)
		}
	}
	return r, nil
}
