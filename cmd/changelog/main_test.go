// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionUp(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		version  semver
		expected semver
		unstable bool
		breaking bool
		minor    bool
	}{
		{
			name:     "unstable major",
			version:  semver{0, 1, 1},
			expected: semver{0, 2, 0},
			unstable: true,
			breaking: true,
		},
		{
			name:     "unstable minor",
			version:  semver{0, 1, 1},
			expected: semver{0, 1, 2},
			unstable: true,
			minor:    true,
		},
		{
			name:     "unstable patch",
			version:  semver{0, 1, 1},
			expected: semver{0, 1, 2},
			unstable: true,
		},
		{
			name:     "stable major",
			version:  semver{1, 1, 1},
			expected: semver{2, 0, 0},
			breaking: true,
		},
		{
			name:     "stable minor",
			version:  semver{1, 1, 1},
			expected: semver{1, 2, 0},
			minor:    true,
		},
		{
			name:     "stable patch",
			version:  semver{1, 1, 1},
			expected: semver{1, 1, 2},
		},
	}

	for _, useCase := range cases {
		t.Run(useCase.name, func(t *testing.T) {
			t.Parallel()
			assert := require.New(t)

			neu := versionUp(useCase.version, useCase.unstable, useCase.breaking, useCase.minor)
			assert.Equal(useCase.expected, neu, useCase.name)
		})
	}
}

func TestParseGitLog(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
	gitLog := []string{
		"5974cb8f96fb6da96a5b917c5f43203daee1b431",
		"fix: correct cheese to be creamy (#43)",
		"* docs(pizza): document how to eat using hands",
		"* fix: correct cheese to be creamy",
		"884d9111c8f62a27c2185c45a1a0211db7277872",
		" (tag: other/v0.1.4, tag: pizza/v0.1.4)",
		"misc: update other (#42)",
	}
	expected := []changelogLine{
		{Text: "* docs(pizza): document how to eat using hands", Hash: "5974cb8f96fb6da96a5b917c5f43203daee1b431"},
		{Text: "* fix: correct cheese to be creamy", Hash: "5974cb8f96fb6da96a5b917c5f43203daee1b431"},
	}

	lines, oldVer, err := parseGitLog("pizza", gitLog)
	assert.NoError(err)
	assert.Equal(semver{0, 1, 4}, oldVer)
	assert.Len(lines, len(expected))

	for i, expect := range expected {
		assert.Equal(expect, lines[i])
	}
}
