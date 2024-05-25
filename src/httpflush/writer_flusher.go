package httpflush

import "net/http"

type writerFlusher interface {
	http.ResponseWriter
	http.Flusher
}
