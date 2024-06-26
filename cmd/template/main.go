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
	Food string `yaml:"food"`
}

var (
	//go:embed config.yml
	raw     []byte
	configs config
	flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger  = &logging.Logger{}
)

func main() {
	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Fatalf("%v", msg)
		}

		if err != nil {
			logger.Error().Log(err.Error())
			os.Exit(1)
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
	_ = flagset.Bool("pizza", true, "pepperoni or mozzarella!. (default TRUE)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	err = yaml.Unmarshal(raw, &configs)
	if err != nil {
		usageExit(err)
	}

	err = template()
}

func usageExit(err error) {
	fmt.Println("description here")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		logger.Error().Log(err.Error())
	}

	os.Exit(1)
}

func template() error {
	return nil
}
