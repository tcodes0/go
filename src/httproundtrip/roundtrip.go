package httproundtrip

import (
	"net/http"

	"github.com/tcodes0/go/src/errutil"
	"github.com/tcodes0/go/src/logging"
)

type Roundtrip struct {
	Transport *http.Transport
	Logger    *logging.Logger
	UserAgent string
}

func (r Roundtrip) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}

	if r.Logger == nil {
		r.Logger = &logging.Logger{}
	}

	r.Logger.Debug().
		Metadata("method", req.Method).
		Metadata("url", req.URL.String()).
		Metadata("headers", req.Header).
		Log("req")

	res, err := r.Transport.RoundTrip(req)

	r.Logger.Debug().
		Metadata("status", res.StatusCode).
		Metadata("length", res.ContentLength).
		Metadata("headers", res.Header).
		Log("res")

	return res, errutil.Wrap(err, "http roundtrip")
}
