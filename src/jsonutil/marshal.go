package jsonutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// MarshalReader marshals data to a reader, panics on error.
func MarshalReader(data any) io.ReadCloser {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return io.NopCloser(bytes.NewReader(b))
}

func MarshalRequest(ctx context.Context, url string, data any) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, "", url, MarshalReader(data))
}
