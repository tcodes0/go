// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

//go:embed header.txt
var licenseHeader string

const (
	golang int = iota + 1
	shell
)

type glob string

func (glob glob) String() string {
	return string(glob)
}

func (glob glob) CommentToken() string {
	switch glob.Kind() {
	default:
		return ""
	case golang:
		return "// "
	case shell:
		return "# "
	}
}

func (glob glob) Kind() int {
	if strings.HasSuffix(glob.String(), ".go") {
		return golang
	}

	if strings.HasSuffix(glob.String(), ".sh") {
		return shell
	}

	return 0
}

type File interface {
	io.ReadCloser
	WriteString(s string) (n int, err error)
}

type OSFiles interface {
	Glob(pattern string) (matches []string, err error)
	OpenFile(name string, flag int, perm os.FileMode) (file File, err error)
	Open(name string) (file File, err error)
	ReadAll(r io.Reader) (b []byte, err error)
}

type osFiles struct{}

var _ OSFiles = (*osFiles)(nil)

func (osf osFiles) Glob(pattern string) (matches []string, err error) {
	//nolint:wrapcheck // test
	return filepath.Glob(pattern)
}

func (osf osFiles) OpenFile(name string, flags int, perm os.FileMode) (file File, err error) {
	//nolint:wrapcheck // test
	return os.OpenFile(name, flags, perm)
}

func (osf osFiles) Open(name string) (file File, err error) {
	//nolint:wrapcheck // test
	return os.Open(name)
}

func (osf osFiles) ReadAll(r io.Reader) (b []byte, err error) {
	//nolint:wrapcheck // test
	return io.ReadAll(r)
}

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs")
	fColor := flagset.Bool("color", false, "colored logging output. default false")
	fGlobs := flagset.String("globs", "", "comma-space separated list of globs to search for files. Required")
	fIgnore := flagset.String("ignore", "", "comma-space separated list of regexes to exclude files by path match. Default empty")
	fDryrun := flagset.Bool("dryrun", false, "do not modify files, just log. Default false")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("%s applies a copyright boilerplate header to files\n", os.Args[0])
		fmt.Printf("ERROR: failed to parse flags: %v", err)
		os.Exit(1)
	}

	if *fGlobs == "" {
		fmt.Printf("%s applies a copyright boilerplate header to files\n", os.Args[0])
		flagset.Usage()
		os.Exit(1)
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)
	globs := strings.Split(*fGlobs, ", ")
	rawRegexps := strings.Split(*fIgnore, ", ")

	if globs == nil {
		logger.Debug().Log("no globs provided")

		return
	}

	ignores := make([]*regexp.Regexp, 0, len(rawRegexps))

	for _, raw := range rawRegexps {
		var reg *regexp.Regexp

		if raw == "" {
			continue
		}

		reg, err = regexp.Compile(raw)
		if err != nil {
			logger.Fatalf("failed to compile regexp %s: %v", raw, err)
		}

		ignores = append(ignores, reg)
	}

	err = boilerplate(*logger, osFiles{}, globs, ignores, licenseHeader, *fDryrun)
	if err != nil {
		logger.Fatalf("failed: %v", err)
	}
}

func boilerplate(
	logger logging.Logger,
	osf OSFiles,
	globs []string,
	ignoreRegexps []*regexp.Regexp,
	header string,
	dryrun bool,
) error {
	for _, fileglob := range globs {
		filePaths, err := osf.Glob(fileglob)
		if err != nil {
			return misc.Wrap(err, "failed to glob files")
		}

		logger.Debug().Logf("glob: '%s', count %d, files: %s", fileglob, len(filePaths), filePaths)

		if len(filePaths) == 0 {
			logger.Warn().Logf("no files matched: %s", fileglob)

			continue
		}

		headerWithComments := ""

	matchesLoop:
		for _, filePath := range filePaths {
			for _, regexp := range ignoreRegexps {
				if regexp.MatchString(filePath) {
					logger.Debug().Logf("skipping %s because ignore '%s' matches", filePath, regexp.String())

					continue matchesLoop
				}
			}

			if headerWithComments == "" {
				headerWithComments = addComments(header, glob(fileglob))
			}

			err, processed := processFile(logger, osf, filePath, fileglob, headerWithComments, dryrun)
			if err != nil && !errors.Is(err, context.Canceled) {
				return misc.Wrapf(err, "failed: %s", filePath)
			}

			if processed {
				logger.Log(filePath)
			}
		}
	}

	return nil
}

func processFile(
	logger logging.Logger,
	osf OSFiles,
	path,
	fileGlob,
	header string,
	dryrun bool,
) (err error, processed bool) {
	hasHeader, content, err := checkForHeader(osf, path, header)
	if err != nil {
		return misc.Wrap(err, "failed to check for header"), false
	}

	if hasHeader {
		logger.Debug().Logf("header already applied: %s", path)

		return nil, false
	}

	if !dryrun {
		err = applyHeader(osf, header, path, content, glob(fileGlob))
		if err != nil {
			return misc.Wrap(err, "failed to apply header"), false
		}
	}

	return nil, true
}

func checkForHeader(osf OSFiles, path, header string) (hasHeader bool, content string, err error) {
	file, err := osf.Open(path)
	if err != nil {
		return false, "", misc.Wrap(err, "failed to open file")
	}

	defer file.Close()

	b, err := osf.ReadAll(file)
	if err != nil {
		return false, "", misc.Wrap(err, "failed to read file")
	}

	content = string(b)

	return strings.Contains(content, header), content, nil
}

func addComments(license string, fileGlob glob) (commentedLicense string) {
	for _, licenseLine := range strings.Split(license, "\n") {
		commentedLicense += fileGlob.CommentToken() + licenseLine + "\n"
	}

	return commentedLicense
}

func applyHeader(osf OSFiles, header, path, content string, glob glob) error {
	file, err := osf.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return misc.Wrap(err, "unable to open file")
	}

	defer file.Close()

	switch glob.Kind() {
	default:
		return fmt.Errorf("unknown kind %d", glob.Kind())
	case golang:
		_, err = file.WriteString(header + "\n" + content)
	case shell:
		shebang, rest, found := strings.Cut(content, "\n")
		if !found {
			return misc.Wrapf(err, "unable to parse %s", path)
		} else {
			_, err = file.WriteString(shebang + "\n" + header + rest)
		}
	}

	if err != nil {
		return misc.Wrap(err, "unable to write to file")
	}

	return nil
}