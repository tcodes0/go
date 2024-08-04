// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:funlen // test
func TestParseGitLog(t *testing.T) {
	t.Parallel()

	head := "4cd0e60 (HEAD -> my-branch) refactor(cmd/pizza): spread ingredients more evenly across"
	minor := "785c362 feat(cmd/pizza): add cheese"
	breaking := "785c362 feat(cmd/pizza)!: add swiss cheese"
	main := "a34ccaf (origin/main, origin/HEAD, main) feat(kitchen): new oven gets hotter!"
	tagUnstable := "78583a3 (tag: pizza/v0.1.1) misc: improve border crunchiness"
	tagStable := "78583a3 (tag: pizza/v1.1.1) misc: improve border crunchiness"
	cases := []struct {
		name      string
		wantOld   string
		wantNew   string
		lines     []string
		wantLines int
	}{
		{
			name:      "unstable major",
			wantOld:   "0.1.1",
			wantNew:   "0.2.0",
			wantLines: 2,
			lines:     []string{head, breaking, main, tagUnstable},
		},
		{
			name:      "unstable minor",
			wantOld:   "0.1.1",
			wantNew:   "0.1.2",
			wantLines: 2,
			lines:     []string{head, minor, main, tagUnstable},
		},
		{
			name:      "unstable patch",
			wantOld:   "0.1.1",
			wantNew:   "0.1.2",
			wantLines: 1,
			lines:     []string{head, main, tagUnstable},
		},
		{
			name:      "stable major",
			wantOld:   "1.1.1",
			wantNew:   "2.0.0",
			wantLines: 2,
			lines:     []string{head, breaking, main, tagStable},
		},
		{
			name:      "stable minor",
			wantOld:   "1.1.1",
			wantNew:   "1.2.0",
			wantLines: 2,
			lines:     []string{head, minor, main, tagStable},
		},
		{
			name:      "stable patch",
			wantOld:   "1.1.1",
			wantNew:   "1.1.2",
			wantLines: 1,
			lines:     []string{head, main, tagStable},
		},
	}

	for _, useCase := range cases {
		t.Run(useCase.name, func(t *testing.T) {
			t.Parallel()
			assert := require.New(t)

			lines, old, neu, err := parseGitLog("pizza", useCase.lines)
			assert.NoError(err, useCase.name)

			assert.Equal(useCase.wantOld, old, useCase.name)
			assert.Equal(useCase.wantNew, neu, useCase.name)
			assert.Len(lines, useCase.wantLines, useCase.name)
		})
	}
}
