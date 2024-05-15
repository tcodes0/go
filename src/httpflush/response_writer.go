package httpflush

import "net/http"

//nolint:unused // used to generate mocks
type responseWriter interface {
	http.ResponseWriter
	http.Flusher
}
