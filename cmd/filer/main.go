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
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

const (
	actionSymlink = "symlink"
	actionRemove  = "remove"
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
	fCommitL := flagset.Bool("commit", false, "apply changes (default: false)")
	fCommitS := flagset.Bool("c", false, "apply changes (default: false)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	fDryrun := !(*fCommitL || *fCommitS)

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

	envVarResolver(configs)

	logger = logging.Create(opts...)
	err = filer(*logger, configs, fDryrun)
}

func readConfig(file string) ([]*config, error) {
	raw := []byte{}

	if file != "" {
		cfgFile, err := os.Open(file)
		if err != nil {
			return nil, misc.Wrap(err, "open")
		}

		raw, err = io.ReadAll(cfgFile)
		if err != nil {
			return nil, misc.Wrap(err, "read")
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
	fmt.Println("execute a config of simple file tasks.")
	fmt.Println("changes no files by default.")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func envVarResolver(configs []*config) {
	envs := map[string]string{}
	envs["$HOME"] = os.Getenv("HOME")

	for _, conf := range configs {
		conf.Input = lo.Map(conf.Input, func(input string, _ int) string {
			return strings.ReplaceAll(input, "$HOME", envs["$HOME"])
		})
	}
}

func filer(logger logging.Logger, configs []*config, dryrun bool) error {
	for _, config := range configs {
		if config == nil {
			continue
		}

		if config.Action == actionSymlink {
			err := symlink(logger, config.Input, dryrun)
			if err != nil {
				return err
			}

			continue
		}

		if config.Action == actionRemove {
			err := remove(logger, config.Input, dryrun)
			if err != nil {
				return err
			}

			continue
		}

		logger.Warn().Logf("ignore: unknown action %s", config.Action)
	}

	if dryrun {
		fmt.Printf("to apply changes run: %s -commit", strings.Join(os.Args, " "))
	}

	return nil
}

func symlink(logger logging.Logger, input []string, dryrun bool) error {
	if len(input) != 1 {
		return fmt.Errorf("symlink: expected 2 inputs got %v", input)
	}

	source := input[0]
	link := input[1]

	_, err := os.Stat(source)
	if err != nil {
		return misc.Wrapf(err, "stat")
	}

	_, err = os.Stat(link)
	if err == nil {
		logger.Warn().Logf("skip: file exists %s", link)

		return nil
	}

	if dryrun {
		_, err = fmt.Printf("symlink %s -> %s\n", link, source)
		if err != nil {
			logger.Error().Logf("println: %v", err)
		}

		return nil
	}

	err = os.Symlink(source, link)
	if err != nil {
		return misc.Wrapf(err, "symlink %s %s", source, link)
	}

	return nil
}

func remove(logger logging.Logger, input []string, dryrun bool) (err error) {
	if len(input) != 1 {
		return fmt.Errorf("remove: expected 1 input got %v", input)
	}

	if dryrun {
		_, err = fmt.Printf("remove %s\n", input[0])
		if err != nil {
			logger.Error().Logf("println: %v", err)
		}

		return nil
	}

	err = os.Remove(input[0])
	if err != nil {
		return misc.Wrap(err, "remove")
	}

	return nil
}
