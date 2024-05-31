package jsonutil_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/jsonutil"
)

func TestUnmarshalReader(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type T struct {
		A string `json:"a"`
	}

	data := &T{A: "a"}
	b, err := json.Marshal(data)
	assert.NoError(err)

	r := io.NopCloser(bytes.NewReader(b))

	out, err := jsonutil.UnmarshalReader[T](r)
	assert.NoError(err)

	assert.Equal(data, out)
}
