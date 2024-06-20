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

	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	Action string   `yaml:"action"`
	Input  []string `yaml:"input"`
}

var flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

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

	fConfig := flagset.String("config", "", "config file path (required)")
	fDryrunL := flagset.Bool("dryrun", false, "do not make changes just print (default false)")
	fDryrunS := flagset.Bool("n", false, "do not make changes just print (default false)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	fDryrun := *fDryrunL || *fDryrunS

	configs, err := readConfig(*fConfig)
	if err != nil {
		usageExit(err)
	}

	if len(configs) == 0 || len(configs) == 1 && configs[0] == nil {
		usageExit(errors.New("empty config"))
	}

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	err = filer(*logger, configs, fDryrun)
}

func readConfig(file string) ([]*config, error) {
	raw := []byte{}

	if file != "" {
		cfgFile, err := os.Open(file)
		if err != nil {
			return nil, misc.Wrapf(err, "open %s", file)
		}

		raw, err = io.ReadAll(cfgFile)
		if err != nil {
			return nil, misc.Wrapf(err, "read %s", file)
		}
	}

	configs := []*config{}

	err := yaml.Unmarshal(raw, &configs)
	if err != nil {
		return nil, misc.Wrap(err, "unmarshal")
	}

	return configs, nil
}

func usageExit(err error) {
	flagset.Usage()
	fmt.Println("execute a config of simple file tasks")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func filer(logger logging.Logger, configs []*config, dryrun bool) error {
	for _, config := range configs {
		if config == nil {
			continue
		}

		if config.Action == "symlink" {
			err := symlink(logger, config.Input[0], config.Input[1], dryrun)
			if err != nil {
				return err
			}

			continue
		}

		logger.Warn().Logf("ignore: unknown action %s", config.Action)
	}

	return nil
}

func symlink(logger logging.Logger, target, link string, dryrun bool) error {
	_, err := os.Stat(target)
	if err != nil {
		return misc.Wrapf(err, "stat")
	}

	_, err = os.Stat(link)
	if err == nil {
		logger.Warn().Logf("skip: symlink %s already exists", link)

		return nil
	}

	if dryrun {
		logger.Logf("symlink %s -> %s", link, target)

		return nil
	}

	err = os.Symlink(target, link)
	if err != nil {
		return misc.Wrapf(err, "symlink %s %s", target, link)
	}

	return nil
}
