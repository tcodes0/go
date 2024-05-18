package httproundtrip

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Roundtrip struct {
	Transport *http.Transport
	Logger    *zerolog.Logger
	UserAgent string
}

func (r Roundtrip) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}

	r.Logger.Debug().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Interface("headers", req.Header).
		Msg("req")

	res, err := r.Transport.RoundTrip(req)

	if r.Logger == nil {
		nop := zerolog.Nop()
		r.Logger = &nop
	}

	r.Logger.Debug().
		Int("status", res.StatusCode).
		Int64("length", res.ContentLength).
		Interface("headers", res.Header).
		Msg("res")

	return res, err
}
