package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/misc"
)

func TestCopyPointed(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type Testcase struct {
		value *any
		name  string
	}

	str := any("abc")
	n := 0
	nPtr := any(&n)
	d := any(struct{ name string }{name: "a"})

	tests := []Testcase{
		{name: "*struct", value: &d},
		{name: "*string", value: &str},
		{name: "nil", value: nil},
		{name: "**int", value: &nPtr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			val := misc.CopyPointed(test.value)

			if test.value != nil {
				assert.Equal(val, *test.value)
			} else {
				assert.Nil(test.value)
			}

			//nolint:gosec // test
			assert.NotEqual(fmt.Sprintf("%p", &val), fmt.Sprintf("%p", &test.value))
		})
	}
}
