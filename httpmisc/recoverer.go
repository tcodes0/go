// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

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

				logger.ErrorData(map[string]any{
					"recover":    msg,
					"stacktrace": debug.Stack(),
				}, "panic")

				http.Error(writer, "{\"error\": \"ERROR\"}", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(writer, req)
	}

	return http.HandlerFunc(middlewareFunc)
}
