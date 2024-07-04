// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package httpmisc

import (
	"net/http"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

// implements http.RoundTripper with debug logging.
type Roundtrip struct {
	Transport *http.Transport
	Logger    *logging.Logger
	UserAgent string
}

var _ http.RoundTripper = (*Roundtrip)(nil)

// executes a roundtrip with debug logging.
func (r Roundtrip) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}

	if r.Logger == nil {
		r.Logger = &logging.Logger{}
	}

	r.Logger.DebugData(map[string]any{
		"method":  req.Method,
		"url":     req.URL.String(),
		"headers": req.Header,
	}, "req")

	res, err := r.Transport.RoundTrip(req)

	r.Logger.DebugData(map[string]any{
		"status":  res.Status,
		"length":  res.ContentLength,
		"headers": res.Header,
	}, "res")

	return res, misc.Wrap(err, "http roundtrip")
}
