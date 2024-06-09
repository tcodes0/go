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

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs")
	fColor := flagset.Bool("color", false, "colored logging output. default false")
	fReport := flagset.Bool("report", false, "do not modify files, error on files missing copyright. Suppress other output. Default false")
	fGlobs := flagset.String("globs", "", "comma-space separated list of globs to search for files. Required")
	fIgnore := flagset.String("ignore", "", "comma-space separated list of regexes to exclude files by path match. Default empty")

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

	if *fReport {
		logger.SetLevel(logging.LError)
	}

	ignores := make([]*regexp.Regexp, 0, len(rawRegexps))

	for _, raw := range rawRegexps {
		var reg *regexp.Regexp

		if raw == "" {
			continue
		}

		reg, err = regexp.Compile(raw)
		if err != nil {
			logger.Fatalf("compiling regexp %s: %v", raw, err)
		}

		ignores = append(ignores, reg)
	}

	err = boilerplate(*logger, globs, ignores, licenseHeader, *fReport)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func boilerplate(logger logging.Logger, globs []string, ignoreRegexps []*regexp.Regexp, header string, report bool) error {
	filesReported := 0

	for _, fileglob := range globs {
		filePaths, err := filepath.Glob(fileglob)
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
				headerWithComments = addComments(header, fileglob)
			}

			reported, err := processFile(logger, filePath, headerWithComments, report)
			if err != nil && !errors.Is(err, context.Canceled) {
				return misc.Wrapf(err, "failed: %s", filePath)
			}

			filesReported += reported
		}
	}

	if filesReported > 0 {
		return fmt.Errorf("files missing copyright header: %d", filesReported)
	}

	return nil
}

func addComments(license, fileGlob string) (commentedLicense string) {
	for _, licenseLine := range strings.Split(license, "\n") {
		if strings.Contains(fileGlob, ".go") {
			commentedLicense += "// " + licenseLine + "\n"
		} else if strings.Contains(fileGlob, ".sh") {
			commentedLicense += "# " + licenseLine + "\n"
		} else {
			panic(fmt.Errorf("unknown file type: %s", fileGlob))
		}
	}

	return commentedLicense
}

func processFile(logger logging.Logger, path, header string, report bool) (reported int, err error) {
	hasHeader, content, err := detectHeader(path, header)
	if err != nil {
		return 0, misc.Wrap(err, "detecting header")
	}

	if hasHeader {
		logger.Debug().Logf("has header: %s", path)

		return 0, nil
	}

	if report {
		logger.Error().Log(path)

		return 1, nil
	}

	err = applyHeader(header, path, content)
	if err != nil {
		return 0, misc.Wrap(err, "appling header")
	}

	logger.Log(path)

	return 0, nil
}

func detectHeader(path, header string) (hasHeader bool, content string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return false, "", misc.Wrap(err, "opening file")
	}

	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return false, "", misc.Wrap(err, "reading file")
	}

	content = string(b)

	return strings.Contains(content, header), content, nil
}

func applyHeader(header, path, content string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return misc.Wrap(err, "opening file for write")
	}

	defer file.Close()

	if strings.Contains(path, ".go") {
		_, err = file.WriteString(header + "\n" + content)
	} else if strings.Contains(path, ".sh") {
		shebang, rest, found := strings.Cut(content, "\n")
		if !found {
			return misc.Wrapf(err, "parsing %s", path)
		} else {
			_, err = file.WriteString(shebang + "\n" + header + rest)
		}
	} else {
		return fmt.Errorf("unknown file type: %s", path)
	}

	if err != nil {
		return misc.Wrap(err, "writing to file")
	}

	return nil
}
