package test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/misc"
	"github.com/tcodes0/go/src/reflectutil"
)

func TestIsNil(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	tests := []struct {
		name  string
		value reflect.Value
		want  bool
	}{
		{
			name:  "nil chan",
			value: reflect.ValueOf(chan int(nil)),
			want:  true,
		},
		{
			name:  "map",
			value: reflect.ValueOf(map[int]int{}),
			want:  false,
		},
		{
			name:  "int pointer",
			value: reflect.ValueOf(misc.PointerTo(33)),
			want:  false,
		},
		{
			name:  "empty string",
			value: reflect.ValueOf(""),
			want:  false,
		},
		{
			name:  "zero",
			value: reflect.ValueOf(0),
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(test.want, reflectutil.IsNil(test.value))
		})
	}
}
