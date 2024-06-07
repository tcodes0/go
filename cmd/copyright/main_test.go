package main

import (
	"bytes"
	_ "embed"
	"io/fs"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/logging"
)

var (
	header = "Copyright Mr. Tester Golang 2021"
	goFile = `package main

func main() {}
`
	goFileWithHeader = `// Copyright Mr. Tester Golang 2021

package main

func main() {}
`
	shFile = `#! /bin/bash

echo "hello"
`
	shFileWithHeader = `#! /bin/bash
# Copyright Mr. Tester Golang 2021

echo "hello"
`
	goTestWithHeader = `// Copyright Mr. Tester Golang 2021
package main

func TestMain() {}
`
)

//nolint:funlen // test
func TestBoilerplate(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	t.Run("two files without header", func(t *testing.T) {
		t.Parallel()
		osf := NewMockOSFiles(t)
		buf := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))

		for _, args := range [][]string{
			{"*.go", "main.go", goFile, goFileWithHeader},
			{"*.sh", "hello.sh", shFile, shFileWithHeader},
		} {
			mockFile := NewMockFile(t)

			osf.Expect().Glob(args[0]).Return([]string{args[1]}, nil).Once()
			osf.Expect().Open(args[1]).Return(mockFile, nil).Once()
			mockFile.Expect().Close().Return(nil).Once()
			osf.Expect().ReadAll(mockFile).Return([]byte(args[2]), nil)

			osf.Expect().OpenFile(args[1], os.O_RDWR|os.O_TRUNC, fs.FileMode(0)).Return(mockFile, nil).Once()
			mockFile.Expect().Close().Return(nil).Once()
			mockFile.Expect().WriteString(args[3]).Return(0, nil).Once()
		}

		err := boilerplate(*logger, osf, []string{"*.go", "*.sh"}, nil, header, false)
		assert.NoError(err)
	})

	t.Run("ignores sh file", func(t *testing.T) {
		t.Parallel()
		osf := NewMockOSFiles(t)
		buf := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))

		for _, args := range [][]string{
			{"*.go", "main.go", goFile, goFileWithHeader},
			{"*.sh", "hello.sh", shFile, shFileWithHeader},
		} {
			mockFile := NewMockFile(t)

			osf.Expect().Glob(args[0]).Return([]string{args[1]}, nil).Once()

			if args[1] == "hello.sh" {
				continue
			}

			osf.Expect().Open(args[1]).Return(mockFile, nil).Once()
			mockFile.Expect().Close().Return(nil).Once()
			osf.Expect().ReadAll(mockFile).Return([]byte(args[2]), nil)

			osf.Expect().OpenFile(args[1], os.O_RDWR|os.O_TRUNC, fs.FileMode(0)).Return(mockFile, nil).Once()
			mockFile.Expect().Close().Return(nil).Once()
			mockFile.Expect().WriteString(args[3]).Return(0, nil).Once()
		}

		err := boilerplate(*logger, osf, []string{"*.go", "*.sh"}, []*regexp.Regexp{regexp.MustCompile("hello*")}, header, false)
		assert.NoError(err)
	})

	t.Run("already has header", func(t *testing.T) {
		t.Parallel()
		osf := NewMockOSFiles(t)
		buf := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))

		args := []string{"*.go", "main_test.go", goTestWithHeader, goTestWithHeader}

		mockFile := NewMockFile(t)

		osf.Expect().Glob(args[0]).Return([]string{args[1]}, nil).Once()

		osf.Expect().Open(args[1]).Return(mockFile, nil).Once()
		mockFile.Expect().Close().Return(nil).Once()
		osf.Expect().ReadAll(mockFile).Return([]byte(args[2]), nil)

		err := boilerplate(*logger, osf, []string{"*.go"}, nil, header, false)
		assert.NoError(err)
	})

	t.Run("invalid glob", func(t *testing.T) {
		t.Parallel()
		osf := NewMockOSFiles(t)
		buf := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))

		args := []string{"*.foobar"}

		osf.Expect().Glob(args[0]).Return(nil, nil).Once()

		err := boilerplate(*logger, osf, []string{"*.foobar"}, nil, header, false)
		assert.NoError(err)
	})

	t.Run("dry run", func(t *testing.T) {
		t.Parallel()
		osf := NewMockOSFiles(t)
		buf := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))

		args := []string{"*.go", "main.go", goFile}

		mockFile := NewMockFile(t)

		osf.Expect().Glob(args[0]).Return([]string{args[1]}, nil).Once()

		osf.Expect().Open(args[1]).Return(mockFile, nil).Once()
		mockFile.Expect().Close().Return(nil).Once()
		osf.Expect().ReadAll(mockFile).Return([]byte(args[2]), nil)

		err := boilerplate(*logger, osf, []string{"*.go"}, nil, header, true)
		assert.NoError(err)
	})
}
