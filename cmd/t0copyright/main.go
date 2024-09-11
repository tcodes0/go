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
	Header  string `yaml:"header"`
	Version string `yaml:"version"`
}

var (
	//go:embed config.yml
	raw           string
	defaultIgnore = `/?mock_.*|.local/.*`
	flagset       = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger        = &logging.Logger{}
	errUsage      = errors.New("see usage")
	errFinal      error
)

//nolint:funlen // main won't lose weight, can't stop growing liMes
func main() {
	defer func() {
		passAway(errFinal)
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
	fFind := flagset.String("check", "", "asterisk glob to check files. (required)")
	fIgnore := flagset.String("ignore", "", fmt.Sprintf("regexp match to ignore. (default %s)", defaultIgnore))
	fVerShort := flagset.Bool("v", false, "print version and exit")
	fVerLong := flagset.Bool("version", false, "print version and exit")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	cfg := config{}

	err = yaml.Unmarshal([]byte(raw), &cfg)
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	if *fVerShort || *fVerLong {
		fmt.Println(cfg.Version)

		return
	}

	if *fFind == "" || (*fComment == "" && *fFix) {
		errFinal = errors.Join(errors.New("missing required flags"), errUsage)

		return
	}

	if fLogLevel == 0 && !*fFix {
		// without -fix and explicit level only print errors
		logger.SetLevel(logging.LError)
	}

	header := ""
	// formatter adds a newline that we don't want
	cfg.Header = strings.TrimSuffix(cfg.Header, "\n")

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

	errFinal = boilerplate(*fFind, header, *fFix, *fShebang, ignore)
}

// Defer from main() very early; the first deferred function will run last.
// Gracefully handles panics and fatal errors. Replaces os.exit(1).
func passAway(fatal error) {
	if msg := recover(); msg != nil {
		logger.Stacktrace(logging.LError, true)
		logger.Fatalf("%v", msg)
	}

	if fatal != nil {
		if errors.Is(fatal, errUsage) || errors.Is(fatal, flag.ErrHelp) {
			usage(fatal)
		}

		logger.Stacktrace(logging.LDebug, true)
		logger.Fatalf("%s", fatal.Error())
	}
}

func usage(err error) {
	if !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
	}

	fmt.Printf(`
recursively finds and reports files missing boilerplate header
-h to see required flags

%s
`, cmd.EnvVarUsage())
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
		logger.Debugf("find: '%s', count %d, files: %s", findName, len(filesFound), filesFound)

		if len(filesFound) == 0 {
			logger.Warnf("no matches: %s", findName)

			continue
		}

		for _, filePath := range filesFound {
			if filePath == "" {
				continue
			}

			if ignore.MatchString(filePath) {
				logger.Debugf("ignored %s", filePath)

				continue
			}

			fileHeader, err := exec.Command("head", filePath).CombinedOutput()
			if err != nil {
				return nil, misc.Wrapf(err, "head %s", filePath)
			}

			// not specific on purpose, header may change if file is from another oss project
			if strings.Contains(string(fileHeader), "Copyright") {
				logger.Debugf("ok: %s", filePath)

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
