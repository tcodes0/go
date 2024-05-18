package httprecoverer

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if msg := recover(); msg != nil && msg != http.ErrAbortHandler {
				logger := zerolog.Ctx(r.Context())
				if logger == nil {
					nop := zerolog.Nop()
					logger = &nop
				}

				logger.Error().Str("panic", fmt.Sprintf("%v", msg)).Send()
				logger.Error().Msgf("stacktrace: %s", string(debug.Stack()))

				http.Error(w, "{\"error\": \"ERROR\"}", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
