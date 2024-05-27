package jsonutil

import (
	"encoding/json"
	"io"

	"github.com/tcodes0/go/src/errutil"
)

// unmarshals a reader to a pointer; closes the reader.
func UnmarshalReader[T any](r io.ReadCloser) (*T, error) {
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errutil.Wrap(err, "reading")
	}

	if len(data) > 0 {
		return UnmarshalBytes[T](data)
	}

	return new(T), nil
}

// unmarshals bytes to a pointer.
func UnmarshalBytes[T any](b []byte) (*T, error) {
	out := new(T)

	if len(b) > 0 {
		err := json.Unmarshal(b, out)
		if err != nil {
			return nil, errutil.Wrap(err, "unmarshalling bytes")
		}
	}

	return out, nil
}
