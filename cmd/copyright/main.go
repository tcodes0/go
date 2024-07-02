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

	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	Header string `yaml:"header"`
}

var (
	//go:embed config.yml
	raw           string
	defaultIgnore = `/?mock_.*|.local/.*`
	flagset       = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger        = &logging.Logger{}
	errUsage      = errors.New("see usage")
)

//nolint:funlen // main won't lose weight, can't stop growing liMes
func main() {
	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Stacktrace(true)
			logger.Fatalf("%v", msg)
		}

		if err != nil {
			if errors.Is(err, errUsage) {
				usage(err)
			}

			logger.Fatalf("%s", err.Error())
		}
	}()

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	fFix := flagset.Bool("fix", false, "write header to files; requires -comment. (default false)")
	fShebang := flagset.Bool("shebang", false, "preserve first line of file, append header after. (default false)")
	fComment := flagset.String("comment", "", "comment token, prepended to header lines. (required if -fix)")
	fFind := flagset.String("find", "", "asterisk glob to find files. (required)")
	fIgnore := flagset.String("ignore", "", fmt.Sprintf("regexp match to ignore. (default %s)", defaultIgnore))

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	cfg := config{}

	err = yaml.Unmarshal([]byte(raw), &cfg)
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	if *fFind == "" || (*fComment == "" && *fFix) {
		err = errors.Join(errors.New("missing required flags"), errUsage)

		return
	}

	if fLogLevel == 0 && !*fFix {
		// without -fix and explicit level only print errors
		logger.SetLevel(logging.LError)
	}

	header := ""

	for _, line := range strings.Split(cfg.Header, "\n") {
		header += *fComment + line + "\n"
	}

	ignore := regexp.MustCompile(defaultIgnore)
	if *fIgnore != "" {
		ignore, err = regexp.Compile(*fIgnore)
		if err != nil {
			return
		}
	}

	err = boilerplate(*fFind, header, *fFix, *fShebang, ignore)
}

func usage(err error) {
	if !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
	}

	fmt.Println()
	fmt.Println("recursively finds and reports files missing boilerplate header")
	fmt.Println("-fix writes the header")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())
}

func boilerplate(findExpr, header string, fix, shebang bool, ignore *regexp.Regexp) error {
	paths, err := filesWithoutHeader(findExpr, ignore)
	if err != nil {
		return misc.Wrapfl(err)
	}

	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths {
		fmt.Println(path)
	}

	if !fix {
		return fmt.Errorf("files missing copyright header: %d", len(paths))
	}

	for _, path := range paths {
		err := fixFile(path, header, shebang)
		if err != nil {
			return misc.Wrapf(err, "fixing %s", path)
		}
	}

	return nil
}

func filesWithoutHeader(findExpr string, ignore *regexp.Regexp) (paths []string, err error) {
	for _, findName := range strings.Split(findExpr, ",") {
		command := exec.Command("find", ".", "-name", findName, "-type", "f")

		findOut, err := command.CombinedOutput()
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

			if ignore.MatchString(filePath) {
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

func fixFile(path, header string, sheB bool) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return misc.Wrapfl(err)
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return misc.Wrapfl(err)
	}

	var newFile string

	if sheB {
		shebang, rest, found := strings.Cut(string(content), "\n")
		if !found {
			return fmt.Errorf("detecting shebang: %s", path)
		}

		newFile = shebang + "\n" + header + rest
	} else {
		newFile = header + "\n" + string(content)
	}

	_, err = file.WriteAt([]byte(newFile), 0)
	if err != nil {
		return misc.Wrapfl(err)
	}

	return nil
}
