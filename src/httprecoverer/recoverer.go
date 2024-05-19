package httprecoverer

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

func Middleware(next http.Handler) http.Handler {
	middlewareFunc := func(writer http.ResponseWriter, req *http.Request) {
		defer func() {
			if msg := recover(); msg != nil && msg != http.ErrAbortHandler {
				logger := zerolog.Ctx(req.Context())
				if logger == nil {
					nop := zerolog.Nop()
					logger = &nop
				}

				logger.Error().Str("panic", fmt.Sprintf("%v", msg)).Send()
				logger.Error().Msgf("stacktrace: %s", string(debug.Stack()))

				http.Error(writer, "{\"error\": \"ERROR\"}", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(writer, req)
	}

	return http.HandlerFunc(middlewareFunc)
}
