// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/cmd/runner"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

const (
	varModule    = "<module>"                // the module passed as input
	varInherit   = "<inherit>"               // copy this from the environment
	varTasksModT = "<task-module-names>"     // all module task names
	varTasksModF = "<task-not-module-names>" // all not module task names
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
		logger.Stacktrace(true)
		logger.Fatalf("%v", msg)
	}

	if fatal != nil {
		if errors.Is(fatal, runner.ErrUsage) || errors.Is(fatal, flag.ErrHelp) {
			usage(fatal)
		}

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
		//nolint:wrapcheck // runner pkg
		return err
	}

	for _, line := range theTask.Exec {
		cmdInput := slices.Concat(strings.Split(line, " "), inputs[1:])
		cmdInput = lo.Map(cmdInput, inputVarMapper(inputs))

		//nolint:gosec // has validation
		command := exec.Command(cmdInput[0], cmdInput[1:]...)

		logger.Info(line)

		if len(theTask.Env) > 0 {
			envs := lo.Map(theTask.Env, envVarMapper(inputs))
			command.Env = append(command.Env, envs...)
		}

		logger.Debugf("env: %s", strings.Join(command.Env, " "))

		stderrBuffer := bytes.Buffer{}
		command.Stderr = &stderrBuffer

		out, err := command.Output()
		if len(out) > 0 {
			fmt.Println(string(out))
		}

		if stderrBuffer.Len() > 0 {
			logger.Info(stderrBuffer.String())
		}

		if err != nil {
			exitErr, ok := (err).(*exec.ExitError)
			if ok && len(exitErr.Stderr) > 0 {
				logger.Errorf("stderr: %s", string(exitErr.Stderr))
			}

			return misc.Wrapf(err, "command '%s'", line)
		}
	}

	return nil
}

func envVarMapper(inputs []string) func(pair string, _ int) string {
	return func(pair string, _ int) string {
		if strings.Contains(pair, varModule) {
			return strings.Replace(pair, varModule, inputs[1], 1)
		}

		if strings.Contains(pair, varInherit) {
			key := strings.Split(pair, "=")[0]

			val, ok := os.LookupEnv(key)
			if !ok {
				logger.Debugf("env value inherited is empty: %s", key)
			}

			return strings.Replace(pair, varInherit, val, 1)
		}

		return pair
	}
}

func inputVarMapper(inputs []string) func(input string, _ int) string {
	return func(input string, _ int) string {
		if strings.Contains(input, varTasksModT) {
			return strings.ReplaceAll(input, varTasksModT, taskNameFilterJoin(cfg.Tasks, func(t *runner.Task, _ int) bool { return t.Module }))
		}

		if strings.Contains(input, varTasksModF) {
			return strings.ReplaceAll(input, varTasksModF, taskNameFilterJoin(cfg.Tasks, func(t *runner.Task, _ int) bool { return !t.Module }))
		}

		if strings.Contains(input, varModule) {
			return strings.ReplaceAll(input, varModule, inputs[1])
		}

		return input
	}
}

func taskNameFilterJoin(tasks []*runner.Task, filterFunc func(t *runner.Task, _ int) bool) string {
	modTasks := lo.Filter(tasks, filterFunc)
	names := lo.Reduce(modTasks, func(agg []string, t *runner.Task, _ int) []string {
		return append(agg, t.Name)
	}, make([]string, 0, len(modTasks)))

	return strings.Join(names, ",")
}
