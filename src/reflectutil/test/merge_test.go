package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/reflectutil"
)

func TestMerge(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type Test struct {
		D *int   `json:"d"`
		A string `json:"a"`
		C []int  `json:"c"`
		B int    `json:"b"`
	}

	type TestCase struct {
		N              string
		Original       *Test
		Target         *Test
		Partial        *Test
		Expected       *Test
		Ignore         []string
		ExpectedIgnore []string
	}

	foo := 69
	bar := 1337

	cases := []TestCase{
		{N: "1", Target: &Test{A: "a", B: 1}, Partial: &Test{A: "b", B: 2}, Expected: &Test{A: "b", B: 2}, Ignore: nil},
		{N: "2", Target: &Test{A: "a", B: 1}, Partial: &Test{A: "", B: 0}, Expected: &Test{A: "a", B: 1}, Ignore: nil},
		{N: "3", Target: &Test{A: "a", B: 1}, Partial: &Test{A: ""}, Expected: &Test{A: "a", B: 1}, Ignore: nil},
		{
			N: "4", Target: &Test{C: []int{1, 2, 3}, D: &foo}, Partial: &Test{C: nil, D: &bar},
			Expected: &Test{C: []int{1, 2, 3}, D: &bar}, Ignore: nil,
		},
		{
			N: "5", Target: &Test{A: "a", B: 1}, Partial: &Test{A: "b", B: 2}, Expected: &Test{A: "a", B: 2},
			Ignore: []string{"A"}, ExpectedIgnore: []string{"A"},
		},
		{N: "6", Target: nil, Partial: &Test{A: "b", B: 2}, Expected: &Test{A: "b", B: 2}, Ignore: nil, ExpectedIgnore: nil},
		{N: "7", Target: &Test{A: "a", B: 1}, Partial: nil, Expected: &Test{A: "a", B: 1}, Ignore: nil, ExpectedIgnore: nil},
	}

	for _, c := range cases {
		out, ignored, err := reflectutil.Merge(c.Target, c.Partial, c.Ignore)
		assert.NoError(err)
		assert.Equal(c.Expected, out, "case %s", c.N)
		assert.Equal(c.ExpectedIgnore, ignored, "case %s", c.N)
	}
}
