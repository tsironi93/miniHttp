package request

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	// "strings"
	"testing"
	"time"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	return n, nil
}

// wrapWithRandomChunks wraps the string in a chunkReader with random chunk size
func wrapWithRandomChunks(s string) io.Reader {
	// create a local rand source
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	chunkSize := r.Intn(10) + 1 // 1..10 bytes per Read

	return &chunkReader{
		data:            s,
		numBytesPerRead: chunkSize,
	}
}

//	func TestRequestLineParse(t *testing.T) {
//		// Test: Good GET Request line
//		r, err := RequestFromReader(wrapWithRandomChunks("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "GET", r.RequestLine.Method)
//		assert.Equal(t, "/", r.RequestLine.RequestTarget)
//		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
//
//		// Test: Good GET Request line with path
//		r, err = RequestFromReader(wrapWithRandomChunks("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "GET", r.RequestLine.Method)
//		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
//		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
//
//		// Test: Invalid number of parts in request line
//		_, err = RequestFromReader(wrapWithRandomChunks("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
//		require.Error(t, err)
//
//		// --- Test: Good GET Request line ---
//		r, err = RequestFromReader(wrapWithRandomChunks(
//			"GET / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
//		))
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "GET", r.RequestLine.Method)
//		assert.Equal(t, "/", r.RequestLine.RequestTarget)
//		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
//
//		// --- Test: Good GET Request line with path ---
//		r, err = RequestFromReader(wrapWithRandomChunks(
//			"GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
//		))
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "GET", r.RequestLine.Method)
//		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
//		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
//
//		// --- Test: Invalid number of parts in request line ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"/coffee HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// === ADDITIONAL TESTS BELOW ===
//
//		// --- Test: Empty request line ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Method missing ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			" / HTTP/1.1\r\nHost: localhost\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Path missing ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"GET  HTTP/1.1\r\nHost: localhost\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Version missing ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"GET / \r\nHost: test\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Wrong HTTP version (missing HTTP/) ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"GET / 1.1\r\nHost: test\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Wrong HTTP version syntax ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"GET / HTTP/x.y\r\nHost: test\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Extra spaces between parts ---
//		r, err = RequestFromReader(wrapWithRandomChunks(
//			"GET     /coffee     HTTP/1.1\r\nHost: localhost\r\n\r\n",
//		))
//		require.NoError(t, err)
//		assert.Equal(t, "GET", r.RequestLine.Method)
//		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
//		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
//
//		// --- Test: Request target containing query ---
//		r, err = RequestFromReader(wrapWithRandomChunks(
//			"GET /search?q=coffee HTTP/1.1\r\nHost: localhost\r\n\r\n",
//		))
//		require.NoError(t, err)
//		assert.Equal(t, "/search?q=coffee", r.RequestLine.RequestTarget)
//
//		// --- Test: Request line too long ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"GET " + strings.Repeat("x", 10000) + " HTTP/1.1\r\nHost: localhost\r\n\r\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: CRLF strictly required ---
//		_, err = RequestFromReader(wrapWithRandomChunks(
//			"GET / HTTP/1.1\nHost: localhost\n\n",
//		))
//		require.Error(t, err)
//
//		// --- Test: Valid POST request ---
//		r, err = RequestFromReader(wrapWithRandomChunks(
//			"POST /form HTTP/1.1\r\nHost: localhost\r\nContent-Length: 5\r\n\r\nhello",
//		))
//		require.NoError(t, err)
//		assert.Equal(t, "POST", r.RequestLine.Method)
//
//		// --- Test: HEAD request ---
//		r, err = RequestFromReader(wrapWithRandomChunks(
//			"HEAD /ping HTTP/1.1\r\nHost: localhost\r\n\r\n",
//		))
//		require.NoError(t, err)
//		assert.Equal(t, "HEAD", r.RequestLine.Method)
//		assert.Equal(t, "/ping", r.RequestLine.RequestTarget)
//	}
//
// func TestHeaderParsing(t *testing.T) {
//
//		// Test: Standard Headers
//		reader := &chunkReader{
//			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
//			numBytesPerRead: 3,
//		}
//		r, err := RequestFromReader(reader)
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "localhost:42069", r.Headers["host"])
//		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
//		assert.Equal(t, "*/*", r.Headers["accept"])
//
//		// Test: Empty Headers
//		reader = &chunkReader{
//			data:            "GET / HTTP/1.1\r\n\r\n",
//			numBytesPerRead: 3,
//		}
//		r, err = RequestFromReader(reader)
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Len(t, r.Headers, 0)
//
//		// Test: Malformed Header
//		reader = &chunkReader{
//			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
//			numBytesPerRead: 3,
//		}
//		r, err = RequestFromReader(reader)
//		require.Error(t, err)
//
//		// Test: Duplicate Headers
//		reader = &chunkReader{
//			data:            "GET / HTTP/1.1\r\nHost: first\r\nHost: second\r\n\r\n",
//			numBytesPerRead: 3,
//		}
//		r, err = RequestFromReader(reader)
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "first, second", r.Headers["host"])
//
//		// Test: Case Insensitive Headers
//		reader = &chunkReader{
//			data:            "GET / HTTP/1.1\r\nHOST: example.com\r\nuser-AGENT: curl/7.81.0\r\n\r\n",
//			numBytesPerRead: 3,
//		}
//		r, err = RequestFromReader(reader)
//		require.NoError(t, err)
//		require.NotNil(t, r)
//		assert.Equal(t, "example.com", r.Headers["host"])
//		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
//
//		// Test: Missing End of Headers
//		reader = &chunkReader{
//			data:            "GET / HTTP/1.1\r\nHost: missingend\r\n",
//			numBytesPerRead: 3,
//		}
//		r, err = RequestFromReader(reader)
//		require.Error(t, err)
//	}
func TestBodyParsing(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Body, 0 reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Body, 0)

	// Test: Empty Body, no reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Body, 0)

	// Test: No Content-Length but Body Exists
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"body without length",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}
