package jsonutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/tcodes0/go/errutil"
)

// marshals data to a reader.
func MarshalReader(data any) (io.ReadCloser, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, errutil.Wrap(err, "marshalling")
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}

// marshals data and creates an http request.
func MarshalRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	b, err := MarshalReader(body)
	if err != nil {
		return nil, errutil.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, b)
	if err != nil {
		return nil, errutil.Wrap(err, "request with context")
	}

	return req, nil
}
