// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bufio"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var (
	//go:embed header.txt
	header   string
	headerGo string
	headerSh string
	flagset  = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
)

func main() {
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs.")
	fColor := flagset.Bool("color", false, "colored logging output. (default false)")
	fFix := flagset.Bool("fix", false, "applies header to files. (default false)")
	fGlobs := flagset.String("globs", "", "comma-space separated list of globs to search for files. (required)")
	fIgnore := flagset.String("ignore", "", "comma-space separated list of regexes to exclude files by path match. (default empty)")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	if *fGlobs == "" {
		usageExit(errors.New("globs required"))
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)
	globs := strings.Split(*fGlobs, ", ")
	rawRegexps := strings.Split(*fIgnore, ", ")

	if *fLogLevel == 0 && !*fFix {
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

	headerLines := strings.Split(header, "\n")

	for _, line := range headerLines {
		headerGo += "// " + line + "\n"
		headerSh += "# " + line + "\n"
	}

	err = boilerplate(*logger, globs, ignores, *fFix)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println("Check and fix missing boilerplate header in files")
	fmt.Println("Without -fix fails if files are missing copyright header and prints files")
	fmt.Println()
	flagset.Usage()

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func boilerplate(logger logging.Logger, globs []string, ignoreRegexps []*regexp.Regexp, fix bool) error {
	var paths []string

	for _, fileglob := range globs {
		cmd := exec.Command("find", ".", "-name", fileglob, "-type", "f")

		output, err := cmd.CombinedOutput()
		if err != nil {
			return misc.Wrapf(err, "find %s", fileglob)
		}

		filePaths := strings.Split(string(output), "\n")
		logger.Debug().Logf("glob: '%s', count %d, files: %s", fileglob, len(filePaths), filePaths)

		if len(filePaths) == 0 {
			logger.Warn().Logf("no files matched: %s", fileglob)

			continue
		}

	filePathsLoop:
		for _, filePath := range filePaths {
			if filePath == "" {
				continue
			}

			for _, regexp := range ignoreRegexps {
				if regexp.MatchString(filePath) {
					logger.Debug().Logf("skip %s ignore '%s'", filePath, regexp.String())

					continue filePathsLoop
				}
			}

			hasHeader, err := detectHeader(filePath)
			if err != nil {
				return misc.Wrap(err, "detecting header")
			}

			if hasHeader {
				logger.Debug().Logf("ok: %s", filePath)

				continue
			}

			paths = append(paths, filePath)
		}
	}

	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths {
		logger.Log(path)
	}

	if !fix {
		return fmt.Errorf("files missing copyright header: %d", len(paths))
	}

	for _, path := range paths {
		err := fixFile(path)
		if err != nil {
			return misc.Wrapf(err, "processing %s", path)
		}
	}

	return nil
}

func detectHeader(path string) (hasHeader bool, err error) {
	file, err := os.Open(path)
	if err != nil {
		return false, misc.Wrap(err, "opening file")
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	head := ""

	for range 2 {
		ok := scanner.Scan()
		if !ok {
			return false, misc.Wrap(scanner.Err(), "scanning")
		}

		head += scanner.Text()
	}

	return strings.Contains(head, "Copyright"), nil
}

func fixFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return misc.Wrap(err, "opening file for write")
	}

	defer file.Close()

	if strings.Contains(path, ".go") {
		_, err = file.WriteAt([]byte(headerGo), 0)
	} else if strings.Contains(path, ".sh") {
		scanner := bufio.NewScanner(file)

		ok := scanner.Scan()
		if !ok {
			return misc.Wrap(scanner.Err(), "scanning")
		}

		// scanner.text contains the first line of the sh file, the shebang
		// write copyright header after
		_, err = file.WriteAt([]byte("\n"+headerSh), int64(len(scanner.Text())))
	} else {
		return fmt.Errorf("unknown file type: %s", path)
	}

	if err != nil {
		return misc.Wrap(err, "writing to file")
	}

	return nil
}
