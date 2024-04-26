package httpflush

import (
	"errors"
	"net/http"
)

var ErrWriterNotFlusher = errors.New("http.ResponseWriter does not implement http.Flusher")

// MaxSize wraps an http.ResponseWriter and flushes every time more than size bytes are written.
// If size is smaller than first write, it will never flush.
type MaxSize struct {
	writtenSinceLastFlush int

	Max    int
	Writer http.ResponseWriter
}

var _ http.ResponseWriter = (*MaxSize)(nil)

func (maxSize *MaxSize) Write(b []byte) (n int, err error) {
	n, err = maxSize.Writer.Write(b)
	maxSize.writtenSinceLastFlush += n

	if maxSize.writtenSinceLastFlush >= maxSize.Max {
		f, ok := maxSize.Writer.(http.Flusher)
		if !ok {
			return 0, ErrWriterNotFlusher
		}

		f.Flush()
		maxSize.writtenSinceLastFlush = 0
	}

	return n, err
}

func (maxSize MaxSize) Header() http.Header {
	return maxSize.Writer.Header()
}

func (maxSize MaxSize) WriteHeader(statusCode int) {
	maxSize.Writer.WriteHeader(statusCode)
}
