// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bytes"
	_ "embed"
	"io/fs"
	"os"
	"regexp"
	"strconv"
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

		err := boilerplate(*logger, osf, []string{"*.go", "*.sh"}, nil, header, false, false)
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

		err := boilerplate(*logger, osf, []string{"*.go", "*.sh"}, []*regexp.Regexp{regexp.MustCompile("hello*")}, header, false, false)
		assert.NoError(err)
	})

	t.Run("invalid glob", func(t *testing.T) {
		t.Parallel()
		osf := NewMockOSFiles(t)
		buf := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))

		args := []string{"*.foobar"}

		osf.Expect().Glob(args[0]).Return(nil, nil).Once()

		err := boilerplate(*logger, osf, []string{"*.foobar"}, nil, header, false, false)
		assert.NoError(err)
	})

	for _, useCase := range [][]string{
		{"already has header", "*.go", "main_test.go", goTestWithHeader, "false", "false", "noError"},
		{"dry run", "*.go", "main.go", goFile, "true", "false", "noError"},
		{"report", "*.go", "main.go", goFile, "false", "true", "Error"},
	} {
		t.Run(useCase[0], func(t *testing.T) {
			t.Parallel()
			osf := NewMockOSFiles(t)
			buf := &bytes.Buffer{}
			logger := logging.Create(logging.OptWriter(buf), logging.OptExit(func(int) {}))
			mockFile := NewMockFile(t)

			osf.Expect().Glob(useCase[1]).Return([]string{useCase[2]}, nil).Once()

			osf.Expect().Open(useCase[2]).Return(mockFile, nil).Once()
			mockFile.Expect().Close().Return(nil).Once()
			osf.Expect().ReadAll(mockFile).Return([]byte(useCase[3]), nil)

			err := boilerplate(*logger, osf, []string{useCase[1]}, nil, header, convBool(useCase[4]), convBool(useCase[5]))
			if useCase[6] == "noError" {
				assert.NoError(err)
			} else {
				assert.Error(err)
			}
		})
	}
}

func convBool(s string) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		panic(err)
	}

	return v
}
