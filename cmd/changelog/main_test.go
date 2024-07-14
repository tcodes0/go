// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGitLog(t *testing.T) {
	t.Parallel()

	head := "4cd0e60 (HEAD -> my-branch) refactor(cmd/pizza): break up into functions"
	minor := "785c362 feat(cmd/pizza): add cheese"
	breaking := "785c362 feat(cmd/pizza)!: add cheese"
	main := "a34ccaf (origin/main, origin/HEAD, main) feat(kitchen): improve oven temperature"
	tagUnstable := "78583a3 (tag: pizza/v0.1.1) misc: rework border crunchiness"
	tagStable := "78583a3 (tag: pizza/v1.1.1) misc: rework border crunchiness"
	cases := []struct {
		name  string
		args  []string
		lines []string
	}{
		{
			name:  "unstable major",
			args:  []string{"0.1.1", "0.2.0", "2"},
			lines: []string{head, breaking, main, tagUnstable},
		},
		{
			name:  "unstable minor",
			args:  []string{"0.1.1", "0.1.2", "2"},
			lines: []string{head, minor, main, tagUnstable},
		},
		{
			name:  "unstable patch",
			args:  []string{"0.1.1", "0.1.2", "1"},
			lines: []string{head, main, tagUnstable},
		},
		{
			name:  "stable major",
			args:  []string{"1.1.1", "2.0.0", "2"},
			lines: []string{head, breaking, main, tagStable},
		},
		{
			name:  "stable minor",
			args:  []string{"1.1.1", "1.2.0", "2"},
			lines: []string{head, minor, main, tagStable},
		},
		{
			name:  "stable patch",
			args:  []string{"1.1.1", "1.1.2", "1"},
			lines: []string{head, main, tagStable},
		},
	}

	for _, useCase := range cases {
		t.Run(useCase.name, func(t *testing.T) {
			t.Parallel()
			assert := require.New(t)

			lines, old, neu, err := parseGitLog("pizza", useCase.lines)
			assert.NoError(err, useCase.name)
			assert.Equal(useCase.args[0], old, useCase.name)
			assert.Equal(useCase.args[1], neu, useCase.name)

			l, _ := strconv.Atoi(useCase.args[2])
			assert.Len(lines, l, useCase.name)
		})
	}
}
