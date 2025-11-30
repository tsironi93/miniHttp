package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders(t *testing.T) {

	t.Run("Valid single header", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")

		n, done, err := h.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", h["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		h := NewHeaders()

		// Pretend parser already has one header
		h["User-Agent"] = "curl/8.0"

		data := []byte("Host: example.com\r\nAccept: */*\r\n\r\n")

		n, done, err := h.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "example.com", h["host"])
		assert.Equal(t, "*/*", h["accept"])
		assert.Equal(t, "curl/8.0", h["User-Agent"])
		assert.Equal(t, len(data)-2, n)
		assert.False(t, done)
	})

	t.Run("Valid done", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("\r\n")

		n, done, err := h.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.True(t, done)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")

		n, done, err := h.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	h := NewHeaders()
	data := []byte("Host: example.com\r\n\r\n")
	n, done, err := h.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "example.com", h["host"]) // key is lowercase
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Test: Key with invalid '@' character (ASCII 64)
	h = NewHeaders()
	data = []byte("Us@er: value\r\n\r\n")
	n, done, err = h.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Value with invalid ASCII character (e.g., 127)
	h = NewHeaders()
	data = []byte("host: value\x7f\r\n\r\n")
	n, done, err = h.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Key with mixed valid and invalid chars
	h = NewHeaders()
	data = []byte("Conten^t-Type: text/html\r\n\r\n")
	n, done, err = h.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/html", h["conten^t-type"])
	assert.False(t, done)

	// Test: Empty key
	h = NewHeaders()
	data = []byte(": value\r\n\r\n")
	n, done, err = h.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple values with same key
	h = NewHeaders()
	data = []byte("Host: localhost69420\r\nSet-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n")
	n, done, err = h.Parse(data)
	require.NoError(t, err)
	require.Equal(t, 106, n)
	assert.False(t, done)
}
