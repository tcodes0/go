package jsonutil

import (
	"encoding/json"
	"io"

	"github.com/tcodes0/go/src/errutil"
)

// Will close reader.
func UnmarshalReader[T any](r io.ReadCloser) (out *T, err error) {
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errutil.Wrap(err, "unmarshal reader: reading")
	}

	if len(data) > 0 {
		return UnmarshalBytes[T](data)
	}

	return out, nil
}

func UnmarshalBytes[T any](b []byte) (out *T, err error) {
	out = new(T)

	if len(b) > 0 {
		err = json.Unmarshal(b, out)
		if err != nil {
			return nil, errutil.Wrap(err, "unmarshal bytes: unmarshalling")
		}
	}

	return out, nil
}
