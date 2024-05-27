package jsonutil_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/jsonutil"
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
