package jsonutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/tcodes0/go/src/errutil"
)

// marshalReader marshals data to a reader, panics on error.
func MarshalReader(data any) io.ReadCloser {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return io.NopCloser(bytes.NewReader(b))
}

func MarshalRequest(ctx context.Context, url string, data any) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "", url, MarshalReader(data))
	if err != nil {
		return nil, errutil.Wrap(err, "request with context")
	}

	return req, nil
}
