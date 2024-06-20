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
	Action string   `yaml:"action"`
	Input  []string `yaml:"input"`
}

var (
	//go:embed config.yml
	raw     string
	configs []config
)

func main() {
	logger := &logging.Logger{}

	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Fatalf("%v", msg)
		}

		if err != nil {
			logger.Error().Logf("%v", err)
			os.Exit(1)
		}
	}()

	err = yaml.Unmarshal([]byte(raw), &configs)
	if err != nil {
		usageExit(err)
	}

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	err = filer(*logger, configs)
}

func usageExit(err error) {
	fmt.Println("execute a config of simple file tasks")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func filer(logger logging.Logger, configs []config) error {
	for _, config := range configs {
		if config.Action == "symlink" {
			err := symlink(logger, config.Input[0], config.Input[1])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func symlink(logger logging.Logger, target, link string) error {
	_, err := os.Stat(target)
	if err != nil {
		return misc.Wrapf(err, "stat")
	}

	_, err = os.Stat(link)
	if err == nil {
		logger.Warn().Logf("skip: symlink %s already exists", link)

		return nil
	}

	err = os.Symlink(target, link)
	if err != nil {
		return misc.Wrapf(err, "symlink %s %s", target, link)
	}

	return nil
}
