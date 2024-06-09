// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package jsonutil_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/jsonutil"
)

func TestMarshalReader(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type T struct {
		A string `json:"a"`
	}

	data := &T{A: "a"}

	r, err := jsonutil.MarshalReader(data)
	assert.NoError(err)
	defer r.Close()

	b, err := io.ReadAll(r)
	assert.NoError(err)

	out := &T{}
	err = json.Unmarshal(b, out)
	assert.NoError(err)

	assert.Equal(data, out)
}

func TestMarshalRequest(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type Dog struct {
		Name string `json:"name"`
	}

	data := &Dog{Name: "bellynha"}
	ctx := context.Background()

	_, err := jsonutil.MarshalRequest(ctx, http.MethodGet, "http://example.com", data)
	assert.NoError(err)
}
