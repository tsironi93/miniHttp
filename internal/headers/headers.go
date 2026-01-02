package headers

import (
	"bytes"
	"errors"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func isValidValue(s string) bool {
	for _, c := range s {
		if c != 9 && (c < 32 || c > 126) {
			return false
		}
	}
	return true
}

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[key]
	return v, ok
}

func isValidKey(s string) bool {
	for _, c := range s {
		switch {
		case c == 33 || (c >= 35 && c <= 39) || c == 42 || c == 45 || c == 46 || c == 124 || c == 126:
		case c >= 48 && c <= 57: // 0-9
		case c >= 65 && c <= 90: // A-Z
		case c >= 94 && c <= 122: // ^-z
		default:
			return false
		}
	}
	return true
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	for len(data) > 0 {
		idx := bytes.Index(data, []byte(crlf))
		if idx == -1 {
			return n, false, nil
		}

		if idx == 0 {
			return n + 2, true, nil
		}

		line := string(data[:idx])
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return n, false, errors.New("invalid header: missing colon")
		}

		if !isValidKey(parts[0]) || !isValidValue(parts[1]) {
			return n, false, errors.New("invalid characters found")
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 && line[colonIdx-1] == ' ' {
			return n, false, errors.New("invalid spacing before colon")
		}

		if key == "" {
			return n, false, errors.New("empty header key")
		}

		v, ok := h[key]
		if ok {
			h[key] = v + ", " + value
		} else {
			h[key] = value
		}

		n += idx + 2
		data = data[idx+2:]
	}

	return n, false, nil
}
