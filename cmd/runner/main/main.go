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
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/cmd/runner"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	Version string         `yaml:"version"`
	Tasks   []*runner.Task `yaml:"tasks"`
}

var (
	cfgFiles   = []string{".t0runnerrc.yml", ".t0runnerrc.yaml"}
	logger     = &logging.Logger{}
	cfg        config
	flagset    = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	errFinal   error
	programVer = "0.2.0"
)

func main() {
	defer func() {
		passAway(errFinal)
	}()

	misc.DotEnv(".env", false /*noisy*/)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	fVerShort := flagset.Bool("v", false, "print version and exit")
	fVerLong := flagset.Bool("version", false, "print version and exit")
	fConfig := flagset.String("config", "", "config file")

	cmdLineArgs, err := parseFlags(os.Args[1:])
	if err != nil {
		errFinal = errors.Join(err, runner.ErrUsage)

		return
	}

	cfgRaw, err := readCfg(fConfig, cfgFiles)
	if err != nil {
		errFinal = errors.Join(err, runner.ErrUsage)

		return
	}

	err = yaml.Unmarshal(cfgRaw, &cfg)
	if err != nil {
		errFinal = errors.Join(err, runner.ErrUsage)

		return
	}

	if *fVerShort || *fVerLong {
		fmt.Println(programVer)

		return
	}

	errFinal = run(cmdLineArgs)
}

// Defer from main() very early; the first deferred function will run last.
// Gracefully handles panics and fatal errors. Replaces os.exit(1).
func passAway(fatal error) {
	if msg := recover(); msg != nil {
		logger.Stacktrace(logging.LError, true)
		logger.Fatalf("%v", msg)
	}

	if fatal != nil {
		if errors.Is(fatal, runner.ErrUsage) || errors.Is(fatal, flag.ErrHelp) {
			usage(fatal)
		}

		logger.Stacktrace(logging.LDebug, true)
		logger.Fatalf("%s", fatal.Error())
	}
}

func usage(incomingErr error) {
	if errors.Is(incomingErr, flag.ErrHelp) {
		fmt.Println()
	}

	packageTasks, repoTasks, builder := []string{}, []string{}, strings.Builder{}

	for _, task := range lo.Filter(cfg.Tasks, func(t *runner.Task, _ int) bool { return t.Package }) {
		line := "./run "
		line += task.Name + "\t"

		packageTasks = append(packageTasks, line)
	}

	for _, task := range lo.Filter(cfg.Tasks, func(t *runner.Task, _ int) bool { return !t.Package }) {
		line := "./run "
		line += task.Name + "\t"

		repoTasks = append(repoTasks, line)
	}

	builder.WriteString(`runner: miscellaneous automation tool
run task:      ./run <task> <args...>
task help:     ./run <task> -h
version:       ./run -v
custom config: ./run -config <file>`)

	if len(packageTasks) > 0 {
		builder.WriteString("\n" + `
package tasks:
` + strings.Join(packageTasks, "\n"))
	}

	if len(repoTasks) > 0 {
		builder.WriteString("\n" + `
repository tasks:
` + strings.Join(repoTasks, "\n") + "\n")
	}

	packages, err := cmd.FindPackages(logger)
	if err != nil {
		fmt.Printf("finding packages: error: %v\n", err)
	}

	if len(packages) > 0 {
		builder.WriteString("\n" + `packages:
- ` + strings.Join(packages, "\n- ") + "\n")
	}

	builder.WriteString("\n" + cmd.EnvVarUsage() + "\n")
	builder.WriteString("\n" + `.env file is checked for environment variables.
see go.doc for config documentation.
default config files: ` + strings.Join(cfgFiles, ", ") + "\n")

	_, _ = fmt.Print(builder.String())
}

func parseFlags(cmdLine []string) (cmdLineArgs []string, err error) {
	err = flagset.Parse(cmdLine)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	skip := false
	for _, cmdL := range cmdLine {
		if skip {
			skip = false

			continue
		}

		if cmdL == "-v" || cmdL == "-version" {
			continue
		}

		if cmdL == "-config" {
			skip = true

			continue
		}

		cmdLineArgs = append(cmdLineArgs, cmdL)
	}

	return cmdLineArgs, nil
}

func readCfg(userCfg *string, defaults []string) (raw []byte, err error) {
	if userCfg != nil && *userCfg != "" {
		raw, err = os.ReadFile(*userCfg)

		return raw, misc.Wrapfl(err)
	}

	for _, defaultCfg := range defaults {
		if _, err := os.Stat(defaultCfg); err == nil {
			raw, err = os.ReadFile(defaultCfg)

			return raw, misc.Wrapfl(err)
		}
	}

	return nil, errors.New("config file not found")
}

// run <task> <package or input1> ...inputs.
func run(inputs []string) error {
	if len(inputs) == 0 {
		return misc.Wrap(runner.ErrUsage, "task is required")
	}

	theTask, found := lo.Find(cfg.Tasks, func(t *runner.Task) bool { return t.Name == inputs[0] })
	if !found {
		taskNames := lo.Map(cfg.Tasks, func(t *runner.Task, _ int) string { return t.Name })

		meant, ok := runner.DidYouMean(inputs[0], taskNames)
		if ok {
			return misc.Wrapf(runner.ErrUsage, "%s: unknown task, %s", inputs[0], meant)
		}

		return misc.Wrapf(runner.ErrUsage, "%s: unknown task", inputs[0])
	}

	err := theTask.Validate(logger, inputs[1:]...)
	if err != nil {
		return misc.Wrapfl(err)
	}

	return misc.Wrapfl(theTask.Execute(logger, cfg.Tasks, inputs...))
}
