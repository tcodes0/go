package test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/misc"
	"github.com/tcodes0/go/src/reflectutil"
)

//nolint:funlen // test
func TestIsNil(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	tests := []struct {
		value any
		name  string
		want  bool
	}{
		{
			name:  "nil chan",
			value: chan int(nil),
			want:  true,
		},
		{
			name:  "map",
			value: map[int]int{},
			want:  false,
		},
		{
			name:  "int pointer",
			value: misc.ToPtr(33),
			want:  false,
		},
		{
			name:  "*int",
			value: (*int)(nil),
			want:  true,
		},
		{
			name:  "zero",
			value: 0,
			want:  false,
		},
		{
			name:  "nil",
			value: nil,
			want:  true,
		},
		{
			name:  "any nil",
			value: any(nil),
			want:  true,
		},
		{
			name:  "reflect value of nil",
			value: reflect.ValueOf(nil),
			want:  true,
		},
		{
			name:  "reflect value of 0",
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

func TestIsZero(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	tests := []struct {
		value any
		name  string
		want  bool
	}{
		{
			name:  "empty string",
			value: "",
			want:  true,
		},
		{
			name:  "nil",
			value: nil,
			want:  false,
		},
		{
			name:  "any nil",
			value: any(nil),
			want:  false,
		},
		{
			name:  "reflect value of nil",
			value: reflect.ValueOf(nil),
			want:  true,
		},
		{
			name:  "reflect value of 0",
			value: reflect.ValueOf(0),
			want:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(test.want, reflectutil.IsZero(test.value))
		})
	}
}
