package httpflush

import (
	"net/http"
)

// Size wraps an http.ResponseWriter and flushes every time more than size bytes are written.
// The http.ResponseWriter must implement http.Flusher.
// If size is smaller than first write, it will never flush.
type Size struct {
	written int

	Size   int
	Writer http.ResponseWriter
}

var _ http.ResponseWriter = (*Size)(nil)

func (s Size) Header() http.Header {
	return s.Writer.Header()
}

func (s *Size) Write(b []byte) (n int, err error) {
	n, err = s.Writer.Write(b)
	s.written += n

	if s.written >= s.Size {
		s.Writer.(http.Flusher).Flush()
		s.written = 0
	}

	return n, err
}

func (s Size) WriteHeader(statusCode int) {
	s.Writer.WriteHeader(statusCode)
}
