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
	//nolint:varnamelen // go is the programming language
	Go int = iota + 1
	Shell
)

type Glob string

func (glob Glob) String() string {
	return string(glob)
}

func (glob Glob) CommentToken() string {
	switch glob.Kind() {
	default:
		return ""
	case Go:
		return "// "
	case Shell:
		return "# "
	}
}

func (glob Glob) Kind() int {
	if strings.HasSuffix(glob.String(), ".go") {
		return Go
	}

	if strings.HasSuffix(glob.String(), ".sh") {
		return Shell
	}

	return 0
}

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, 5 is fatal")
	fColor := flagset.Bool("color", false, "colored logging output. default false")
	fGlobs := flagset.String("globs", "", "comma-space separated list of globs to search for files. Default empty")
	fIgnore := flagset.String("ignore", "", "comma-space separated list of regexes to exclude files by path match. Default empty")
	fDryrun := flagset.Bool("dryrun", false, "do not modify files, only log what would be done. Default false")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("ERR: failed to parse flags: %v", err)
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

	err = CopyrightBoilerplate(*logger, globs, ignores, *fDryrun)
	if err != nil {
		logger.Fatalf("failed: %v", err)
	}
}

func CopyrightBoilerplate(logger logging.Logger, globs []string, ignoreRegexps []*regexp.Regexp, dryrun bool) error {
	for _, glob := range globs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return misc.Wrap(err, "failed to glob files")
		}

		logger.Debug().Logf("glob: '%s', count %d, files: %s", glob, len(matches), matches)

		if len(matches) == 0 {
			logger.Warn().Logf("no files matched: %s", glob)

			continue
		}

		headerWithComments := ""
		errChan := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())

	matchesLoop:
		for _, match := range matches {
			for _, regexp := range ignoreRegexps {
				if regexp.MatchString(match) {
					logger.Debug().Logf("skipping %s because ignore '%s' matches", match, regexp.String())

					continue matchesLoop
				}
			}

			if headerWithComments == "" {
				headerWithComments = addComments(licenseHeader, Glob(glob))
			}

			go func() {
				//nolint:govet // scope
				err, processed := processFile(ctx, logger, match, glob, headerWithComments, dryrun)
				if !errors.Is(err, context.Canceled) {
					errChan <- misc.Wrapf(err, "failed: %s", match)
					cancel()
				}

				if processed {
					logger.Log(match)
				}
			}()
		}

		err = <-errChan
		if err != nil {
			return misc.Wrap(err, "failed to process file")
		}
	}

	return nil
}

func processFile(ctx context.Context, logger logging.Logger, path, glob, header string, dryrun bool) (err error, processed bool) {
	if ctx.Err() != nil {
		return misc.Wrap(ctx.Err(), "context cancelled"), false
	}

	hasHeader, content, err := checkForHeader(path, header)
	if err != nil {
		return misc.Wrap(err, "failed to check for header"), false
	}

	if hasHeader {
		logger.Debug().Logf("header already applied: %s", path)

		return nil, false
	}

	if !dryrun {
		err = applyHeader(header, path, content, Glob(glob))
		if err != nil {
			return misc.Wrap(err, "failed to apply header"), false
		}
	}

	return nil, true
}

func checkForHeader(path, header string) (hasHeader bool, content string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return false, "", misc.Wrap(err, "failed to open file")
	}

	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return false, "", misc.Wrap(err, "failed to read file")
	}

	content = string(b)

	return strings.Contains(content, header), content, nil
}

func addComments(license string, glob Glob) (commentedLicense string) {
	for _, licenseLine := range strings.Split(license, "\n") {
		commentedLicense += glob.CommentToken() + licenseLine + "\n"
	}

	return commentedLicense
}

func applyHeader(header, path, content string, glob Glob) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return misc.Wrap(err, "unable to open file")
	}

	defer file.Close()

	switch glob.Kind() {
	default:
		return fmt.Errorf("unknown kind %d", glob.Kind())
	case Go:
		_, err = file.WriteString(header + "\n" + content)
	case Shell:
		shebang, rest, found := strings.Cut(content, "\n")
		if !found {
			_, err = file.WriteString(header + "\n" + shebang + "\n" + rest)
		} else {
			_, err = file.WriteString(shebang + "\n" + header + rest)
		}
	}

	if err != nil {
		return misc.Wrap(err, "unable to write to file")
	}

	return nil
}
