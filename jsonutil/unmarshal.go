// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package jsonutil

import (
	"encoding/json"
	"io"

	"github.com/tcodes0/go/misc"
)

// unmarshals a reader to a pointer; does not close the reader.
func UnmarshalReader[T any](r io.ReadCloser) (*T, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, misc.Wrap(err, "reading")
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
			return nil, misc.Wrap(err, "unmarshalling bytes")
		}
	}

	return out, nil
}
