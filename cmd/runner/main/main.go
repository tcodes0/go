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
	//go:embed config.yml
	rawConfig string
	logger    = &logging.Logger{}
	cfg       config
	flagset   = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	errFinal  error
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

	err := yaml.Unmarshal([]byte(rawConfig), &cfg)
	if err != nil {
		errFinal = errors.Join(err, runner.ErrUsage)

		return
	}

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		errFinal = errors.Join(err, runner.ErrUsage)

		return
	}

	if *fVerShort || *fVerLong {
		fmt.Println(cfg.Version)

		return
	}

	errFinal = run(os.Args[1:]...)
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

func usage(err error) {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println()
	}

	moduleTasks, repoTasks := []string{}, []string{}

	for _, task := range lo.Filter(cfg.Tasks, func(t *runner.Task, _ int) bool { return t.Module }) {
		line := "./run "
		line += task.Name + "\t"

		moduleTasks = append(moduleTasks, line)
	}

	for _, task := range lo.Filter(cfg.Tasks, func(t *runner.Task, _ int) bool { return !t.Module }) {
		line := "./run "
		line += task.Name + "\t"

		repoTasks = append(repoTasks, line)
	}

	modules, e := cmd.FindModules(logger)
	if e != nil {
		fmt.Printf("finding modules: error: %v\n", e)
	}

	fmt.Printf(`runner: miscellaneous automation tool
run task:  ./run <task> <module?> <other args?>
task help: ./run <task> -h

module tasks:
%s

repository tasks:
%s

modules:
- %s

%s
.env file is used
`, strings.Join(moduleTasks, "\n"), strings.Join(repoTasks, "\n"), strings.Join(modules, "\n- "), cmd.EnvVarUsage())
}

// run <task> <module or input1> ...inputs.
func run(inputs ...string) error {
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

	return misc.Wrapfl(theTask.Execute(logger, cfg.Tasks, inputs[1:]...))
}
