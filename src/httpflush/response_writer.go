package httpflush

import "net/http"

// nolint
type responseWriter interface {
	http.ResponseWriter
	http.Flusher
}
