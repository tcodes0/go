// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/misc"
)

//nolint:funlen // test
func TestMerge(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type Wombat struct {
		Age      *int
		Nickname string
		ID       []int
		Address  int
	}

	type testCase struct {
		name           string
		base           *Wombat
		partial        *Wombat
		expected       *Wombat
		ignore         []string
		expectedIgnore []string
	}

	tests := []testCase{
		{
			name: "field by field", base: &Wombat{Nickname: "sushi", Address: 1}, partial: &Wombat{Nickname: "bib", Address: 2},
			expected: &Wombat{Nickname: "bib", Address: 2}, ignore: nil,
		},
		{
			name: "zero fields", base: &Wombat{Nickname: "sushi", Address: 1}, partial: &Wombat{Nickname: "", Address: 0},
			expected: &Wombat{Nickname: "sushi", Address: 1}, ignore: nil,
		},
		{
			name: "missing field", base: &Wombat{Nickname: "sushi", Address: 1}, partial: &Wombat{Nickname: ""},
			expected: &Wombat{Nickname: "sushi", Address: 1}, ignore: nil,
		},
		{
			name: "nil field", base: &Wombat{ID: []int{1, 2, 3}, Age: misc.ToPtr(12)}, partial: &Wombat{ID: nil, Age: misc.ToPtr(15)},
			expected: &Wombat{ID: []int{1, 2, 3}, Age: misc.ToPtr(15)}, ignore: nil,
		},
		{
			name: "ignored fields", base: &Wombat{Nickname: "sushi", Address: 1}, partial: &Wombat{Nickname: "bib", Address: 2},
			expected: &Wombat{Nickname: "sushi", Address: 2},
			ignore:   []string{"Nickname"}, expectedIgnore: []string{"Nickname"},
		},
		{
			name: "nil base", base: nil, partial: &Wombat{Nickname: "bib", Address: 2}, expected: &Wombat{Nickname: "bib", Address: 2},
			ignore: nil, expectedIgnore: nil,
		},
		{
			name: "nil partial", base: &Wombat{Nickname: "sushi", Address: 1}, partial: nil, expected: &Wombat{Nickname: "sushi", Address: 1},
			ignore: nil, expectedIgnore: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			out, ignored, err := misc.Merge(test.base, test.partial, test.ignore)
			assert.NoError(err)
			assert.Equal(test.expected, out)
			assert.Equal(test.expectedIgnore, ignored)
		})
	}
}
