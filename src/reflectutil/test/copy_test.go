package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/reflectutil"
)

func TestCopyPointed(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type Test struct {
		A string `json:"a"`
	}

	test := &Test{A: "a"}
	val := reflectutil.CopyPointed(test)
	assert.Equal(*test, val)
	assert.NotEqual(fmt.Sprintf("%p", &val), fmt.Sprintf("%p", test))
	val.A = "b"
	assert.NotEqual(*test, val)
}

func TestCopyOf(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type Test struct {
		A string `json:"a"`
	}

	test := Test{A: "a"}
	val := reflectutil.CopyOf(test)
	assert.Equal(test, val)
	assert.NotEqual(fmt.Sprintf("%p", &val), fmt.Sprintf("%p", &test))
}
