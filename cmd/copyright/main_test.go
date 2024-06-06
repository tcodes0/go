package main

import (
	"bytes"
	_ "embed"
	"io"
	"io/fs"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/logging"
)

//go:embed testdata/header.txt
var header string

//nolint:funlen // test
func TestBoilerplate(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type testFile struct {
		name    string
		expect  string
		glob    string
		globbed bool
		opened  bool
		written bool
	}

	useCases := []struct {
		assertErr func(err error, args ...any)
		name      string
		globs     []string
		ignores   []*regexp.Regexp
		files     []*testFile
		logs      bytes.Buffer
		dryrun    bool
	}{
		{
			logs:      bytes.Buffer{},
			globs:     []string{"*.go", "*.sh"},
			ignores:   []*regexp.Regexp{regexp.MustCompile(`.*ignored.*`)},
			dryrun:    false,
			assertErr: assert.NoError,
			files: []*testFile{
				{
					name: "main.go", globbed: true, opened: true, written: true, glob: "*.go",
					expect: `// Copyright MR. Tester Golang 2021

package main

func main() {}
`,
				},
				{
					name: "hello.sh", globbed: true, opened: true, written: true, glob: "*.sh",
					expect: `#! /bin/bash
# Copyright MR. Tester Golang 2021

echo "hello"
`,
				},
			},
		},
	}

	for _, useCase := range useCases {
		osf := NewMockOSFiles(t)

		//nolint:gosec // test
		logger := logging.Create(logging.OptWriter(&useCase.logs), logging.OptExit(func(int) {}))

		for _, file := range useCase.files {
			mockFile := NewMockFile(t)

			f, err := os.Open("testdata/" + file.name)
			assert.NoError(err)

			b, err := io.ReadAll(f)
			assert.NoError(err)

			content := string(b)

			if file.globbed {
				osf.Expect().Glob(file.glob).Return([]string{file.name}, nil).Once()
			}

			if file.opened {
				osf.Expect().Open(file.name).Return(mockFile, nil).Once()
				mockFile.Expect().Close().Return(nil).Once()
				osf.Expect().ReadAll(mockFile).Return([]byte(content), nil)
			}

			if file.written {
				osf.Expect().OpenFile(file.name, os.O_RDWR|os.O_TRUNC, fs.FileMode(0)).Return(mockFile, nil).Once()
				mockFile.Expect().Close().Return(nil).Once()
				mockFile.Expect().WriteString(file.expect).Return(0, nil).Once()
			}
		}

		err := boilerplate(*logger, osf, useCase.globs, useCase.ignores, header, useCase.dryrun)
		useCase.assertErr(err)
	}
}
