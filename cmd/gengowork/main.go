// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

func main() {
	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	err = genGoWork(*logger)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("generates go.work file")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func genGoWork(logger logging.Logger) error {
	version, err := parseGoVersion()
	if err != nil {
		return misc.Wrap(err, "parseGoVersion")
	}

	mods, err := findModules(logger)
	if err != nil {
		return misc.Wrap(err, "FindModules")
	}

	format := `// generated do not edit.
go %s

use (
	.
	%s
)`
	prettyMods := strings.Join(mods, "\n\t")
	newFile := fmt.Sprintf(format, version, prettyMods)

	err = cmd.WriteFile("go.work", []byte(newFile))
	if err != nil {
		return misc.Wrap(err, "WriteFile")
	}

	return nil
}

func parseGoVersion() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", misc.Wrap(err, "opening")
	}

	scanner := bufio.NewScanner(file)
	goVersion := regexp.MustCompile(`go (\d+\.\d+)`)

	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return "", misc.Wrap(err, "scanning")
		}

		line := scanner.Text()

		if goVersion.MatchString(line) {
			return goVersion.FindStringSubmatch(line)[1], nil
		}
	}

	return "", errors.New("parsing")
}

func findModules(logger logging.Logger) ([]string, error) {
	modules, err := cmd.FindModules(logger)
	if err != nil {
		return nil, misc.Wrap(err, "FindModules")
	}

	out := make([]string, 0, len(modules))
	regexpPathHasSlash := regexp.MustCompile(`(.*\w)/.*`)

	for _, mod := range modules {
		if regexpPathHasSlash.MatchString(mod) {
			out = append(out, regexpPathHasSlash.FindStringSubmatch(mod)[1])
		} else {
			out = append(out, mod)
		}
	}

	out = lo.Uniq(out)

	return out, nil
}
