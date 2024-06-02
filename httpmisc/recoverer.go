package httpmisc

import (
	"net/http"
	"runtime/debug"

	"github.com/tcodes0/go/logging"
)

// a middleware that recovers from panics.
func Recoverer(next http.Handler) http.Handler {
	middlewareFunc := func(writer http.ResponseWriter, req *http.Request) {
		//nolint:contextcheck // context in scope
		defer func() {
			if msg := recover(); msg != nil && msg != http.ErrAbortHandler {
				logger := logging.FromContext(req.Context())

				logger.Error().
					Metadata("recover", msg).
					Metadata("stacktrace", debug.Stack()).
					Log("panic")

				http.Error(writer, "{\"error\": \"ERROR\"}", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(writer, req)
	}

	return http.HandlerFunc(middlewareFunc)
}
