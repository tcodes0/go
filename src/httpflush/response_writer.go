package httpflush

import "net/http"

// used to generate mocks only
// nolint
type responseWriter interface {
	http.ResponseWriter
	http.Flusher
}
