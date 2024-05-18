package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/reflectutil"
)

func TestCopyPtrToVal(t *testing.T) {
	assert := require.New(t)

	type Test struct {
		A string `json:"a"`
	}

	test := &Test{A: "a"}
	val := reflectutil.CopyPointerToValue(test)
	assert.Equal(*test, val)
	assert.NotEqual(fmt.Sprintf("%p", &val), fmt.Sprintf("%p", test))
	val.A = "b"
	assert.NotEqual(*test, val)
}

func TestCopyVal(t *testing.T) {
	assert := require.New(t)

	type Test struct {
		A string `json:"a"`
	}

	test := Test{A: "a"}
	val := reflectutil.CopyValue(test)
	assert.Equal(test, val)
	assert.NotEqual(fmt.Sprintf("%p", &val), fmt.Sprintf("%p", &test))
}
