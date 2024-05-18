package test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/jsonutil"
)

func TestMarshalReader(t *testing.T) {
	assert := require.New(t)

	type T struct {
		A string `json:"a"`
	}

	data := &T{A: "a"}
	r := jsonutil.MarshalReader(data)
	defer r.Close()

	b, err := io.ReadAll(r)
	assert.NoError(err)

	out := &T{}
	err = json.Unmarshal(b, out)
	assert.NoError(err)

	assert.Equal(data, out)
}
