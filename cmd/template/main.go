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
	"log"
	"os"

	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	Food    string `yaml:"food"`
	Version string `yaml:"version"`
}

var (
	//go:embed config.yml
	raw      []byte
	configs  config
	flagset  = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger   = &logging.Logger{}
	errUsage = errors.New("see usage")
	errFinal error
)

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
	_ = flagset.Bool("pizza", true, "pepperoni or mozzarella!")
	fVerShort := flagset.Bool("v", false, "print version and exit")
	fVerLong := flagset.Bool("version", false, "print version and exit")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	err = yaml.Unmarshal(raw, &configs)
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	if *fVerShort || *fVerLong {
		fmt.Println(configs.Version)

		return
	}

	errFinal = template()
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
Template

%s
`, cmd.EnvVarUsage())
}

func template() error {
	return nil
}
