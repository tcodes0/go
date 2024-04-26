package httpflush

import "net/http"

type ResponseWriter interface {
	http.ResponseWriter
	Flush()
}
