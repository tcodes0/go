// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	Ignore    *regexp.Regexp
	Header    string `yaml:"header"`
	HeaderGo  string
	HeaderSh  string
	CommentGo string `yaml:"commentGo"`
	CommentSh string `yaml:"commentSh"`
	IgnoreRaw string `yaml:"ignoreRaw"`
	FindNames string `yaml:"findNames"`
}

var (
	//go:embed config.yml
	raw     string
	configs config
	flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
)

func main() {
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs.")
	fColor := flagset.Bool("color", false, "colored logging output. (default false)")
	fFix := flagset.Bool("fix", false, "applies header to files. (default false)")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	err = yaml.Unmarshal([]byte(raw), &configs)
	if err != nil {
		usageExit(err)
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	if *fLogLevel == 0 && !*fFix {
		// withut -fix and explicit level only print errors
		logger.SetLevel(logging.LError)
	}

	headerLines := strings.Split(configs.Header, "\n")

	for _, line := range headerLines {
		configs.HeaderGo += configs.CommentGo + line + "\n"
		configs.HeaderSh += configs.CommentSh + line + "\n"
	}

	configs.Ignore = regexp.MustCompile(configs.IgnoreRaw)

	err = boilerplate(*logger, *fFix)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("Check and fix missing boilerplate header in files")
	fmt.Println("Without -fix fails if files are missing copyright header and prints files")
	fmt.Println()

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func boilerplate(logger logging.Logger, fix bool) error {
	paths, err := filesWithoutHeader(logger)
	if err != nil {
		return misc.Wrap(err, "finding")
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
			return misc.Wrapf(err, "fixing %s", path)
		}
	}

	return nil
}

func filesWithoutHeader(logger logging.Logger) (paths []string, err error) {
	for _, findName := range strings.Split(configs.FindNames, ",") {
		cmd := exec.Command("find", ".", "-name", findName, "-type", "f")

		findOut, err := cmd.CombinedOutput()
		if err != nil {
			return nil, misc.Wrapf(err, "find %s", findName)
		}

		filesFound := strings.Split(string(findOut), "\n")
		logger.Debug().Logf("find: '%s', count %d, files: %s", findName, len(filesFound), filesFound)

		if len(filesFound) == 0 {
			logger.Warn().Logf("no matches: %s", findName)

			continue
		}

		for _, filePath := range filesFound {
			if filePath == "" {
				continue
			}

			if configs.Ignore.MatchString(filePath) {
				logger.Debug().Logf("ignored %s", filePath)

				continue
			}

			fileHeader, err := exec.Command("head", filePath).CombinedOutput()
			if err != nil {
				return nil, misc.Wrapf(err, "head %s", filePath)
			}

			// not specific on purpose, header may change if file is from another oss project
			if strings.Contains(string(fileHeader), "Copyright") {
				logger.Debug().Logf("ok: %s", filePath)

				continue
			}

			paths = append(paths, filePath)
		}
	}

	return paths, nil
}

func fixFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return misc.Wrap(err, "opening")
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return misc.Wrap(err, "reading")
	}

	var newFile string

	if strings.Contains(path, ".go") {
		newFile = configs.HeaderGo + "\n" + string(content)
	} else if strings.Contains(path, ".sh") {
		shebang, rest, found := strings.Cut(string(content), "\n")
		if !found {
			return fmt.Errorf("detecting shebang: %s", path)
		}

		newFile = shebang + "\n" + configs.HeaderSh + rest
	} else {
		return fmt.Errorf("unknown file type: %s", path)
	}

	_, err = file.WriteAt([]byte(newFile), 0)
	if err != nil {
		return misc.Wrap(err, "writing")
	}

	return nil
}
