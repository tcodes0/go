package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:funlen // obese main, need a fix
func TestParseGitLog(t *testing.T) {
	// t.Parallel()

	head := "4cd0e60 (HEAD -> my-branch) refactor(cmd/pizza): break up into functions"
	minor := "785c362 feat(cmd/pizza): add cheese"
	breaking := "785c362 feat(cmd/pizza)!: add cheese"
	main := "a34ccaf (origin/main, origin/HEAD, main) feat(kitchen): improve oven temperature"
	tagUnstable := "78583a3 (tag: pizza/v0.1.1) misc: rework border crunchiness"
	tagStable := "78583a3 (tag: pizza/v1.1.1) misc: rework border crunchiness"
	cases := []struct {
		name   string
		module string
		old    string
		new    string
		lines  []string
		len    int
	}{
		{
			name:   "unstable minor, tag behind main",
			module: "pizza",
			old:    "0.1.1",
			new:    "0.1.2",
			len:    2,
			lines:  []string{head, minor, main, tagUnstable},
		},
		{
			name:   "unstable major, tag behind main",
			module: "pizza",
			old:    "0.1.1",
			new:    "0.2.0",
			len:    2,
			lines:  []string{head, breaking, main, tagUnstable},
		},
		{
			name:   "stable minor, tag behind main",
			module: "pizza",
			old:    "1.1.1",
			new:    "1.2.0",
			len:    2,
			lines:  []string{head, minor, main, tagStable},
		},
		{
			name:   "stable major, tag behind main",
			module: "pizza",
			old:    "1.1.1",
			new:    "2.0.0",
			len:    2,
			lines:  []string{head, breaking, main, tagStable},
		},
		{
			name:   "stable patch, tag ahead of main",
			module: "pizza",
			old:    "1.1.1",
			new:    "1.1.2",
			len:    1,
			lines:  []string{head, tagStable, main},
		},
	}

	for _, useCase := range cases {
		t.Run(useCase.name, func(t *testing.T) {
			// t.Parallel()
			assert := require.New(t)

			lines, old, neu := parseGitLog(useCase.module, useCase.lines)
			assert.Equal(useCase.old, old, useCase.name)
			assert.Equal(useCase.new, neu, useCase.name)
			assert.Len(lines, useCase.len, useCase.name)
		})
	}
}
