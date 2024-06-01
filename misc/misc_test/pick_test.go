package misc_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/misc"
)

//nolint:funlen //test
func TestPickValid(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	num := 23

	tests := []struct {
		want   any
		name   string
		values []any
	}{
		{
			name:   "nil and nil",
			values: []any{nil, nil},
			want:   nil,
		},
		{
			name:   "empty string, string",
			values: []any{"", "abc"},
			want:   "abc",
		},
		{
			name:   "0 0",
			values: []any{0, 0},
			want:   0,
		},
		{
			name:   "0 nil",
			values: []any{0, nil},
			want:   nil,
		},
		{
			name:   "empty string 0 0",
			values: []any{"", 0, 0},
			want:   0,
		},
		{
			name:   "nil map",
			values: []any{nil, map[int]int{}},
			want:   map[int]int{},
		},
		{
			name:   "**int *int",
			values: []any{misc.ToPtr(misc.ToPtr(num)), misc.ToPtr(num), nil},
			want:   misc.ToPtr(misc.ToPtr(num)),
		},
		{
			name:   "*int(nil) *int",
			values: []any{(*int)(nil), misc.ToPtr(num)},
			want:   misc.ToPtr(num),
		},
		{
			name:   "empty call",
			values: []any{},
			want:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(test.want, misc.PickValid(test.values...))
		})
	}
}
