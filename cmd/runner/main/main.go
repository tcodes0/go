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

var (
	//go:embed config.yml
	config string
	tasks  []*runner.Task
	logger = &logging.Logger{}
)

func main() {
	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Fatalf("%v", msg)
		}

		if err != nil {
			if errors.Is(err, runner.ErrUsage) {
				usage(err)
			}

			logger.Fatalf("%s", err.Error())
		}
	}()

	misc.DotEnv(".env", false /*noisy*/)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)

	err = yaml.Unmarshal([]byte(config), &tasks)
	if err != nil {
		err = errors.Join(err, runner.ErrUsage)

		return
	}

	err = run(*logger, os.Args[1:]...)
}

func usage(err error) {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println()
	}

	fmt.Println("miscellaneous automation tool")
	fmt.Println("usage: ./run <task> <module?> <other args?> (run task)")
	fmt.Println("usage: ./run <task> -h (task help)")
	fmt.Println()
	fmt.Println("module tasks:")

	for _, task := range lo.Filter(tasks, func(t *runner.Task, _ int) bool { return t.Module }) {
		line := "./run "
		line += task.Name + "\t"

		fmt.Println(line)
	}

	fmt.Println()
	fmt.Println("repository tasks:")

	for _, task := range lo.Filter(tasks, func(t *runner.Task, _ int) bool { return !t.Module }) {
		line := "./run "
		line += task.Name + "\t"

		fmt.Println(line)
	}

	modules, e := cmd.FindModules(logging.Logger{})
	if e != nil {
		fmt.Printf("finding modules: error: %v\n", e)
	}

	fmt.Println()
	fmt.Println("modules:")
	fmt.Println("- all (iterate all modules)")
	fmt.Println("- " + strings.Join(modules, "\n- "))
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())
	fmt.Println(".env file is used")
}

// run <task> <module or input1> ...inputs.
func run(logger logging.Logger, inputs ...string) error {
	if len(inputs) == 0 {
		return misc.Wrap(runner.ErrUsage, "task is required")
	}

	theTask, found := lo.Find(tasks, func(t *runner.Task) bool { return t.Name == inputs[0] })
	if !found {
		runner.DidYouMean(inputs[0])

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

		logger.Log(line)

		if len(theTask.Env) > 0 {
			envs := lo.Map(theTask.Env, envVarMapper(logger, inputs))
			command.Env = append(command.Env, envs...)
		}

		logger.Debug().Log("env: " + strings.Join(command.Env, " "))

		stderrBuffer := bytes.Buffer{}
		command.Stderr = &stderrBuffer

		out, err := command.Output()
		if len(out) > 0 {
			fmt.Println(string(out))
		}

		if stderrBuffer.Len() > 0 {
			logger.Log(stderrBuffer.String())
		}

		if err != nil {
			exitErr, ok := (err).(*exec.ExitError)
			if ok && len(exitErr.Stderr) > 0 {
				logger.Error().Log("stderr: " + string(exitErr.Stderr))
			}

			return misc.Wrapf(err, "command '%s'", line)
		}
	}

	return nil
}

func envVarMapper(logger logging.Logger, inputs []string) func(pair string, _ int) string {
	return func(pair string, _ int) string {
		if strings.Contains(pair, varModule) {
			return strings.Replace(pair, varModule, inputs[1], 1)
		}

		if strings.Contains(pair, varInherit) {
			key := strings.Split(pair, "=")[0]

			val, ok := os.LookupEnv(key)
			if !ok {
				logger.Debug().Logf("env value inherited is empty: " + key)
			}

			return strings.Replace(pair, varInherit, val, 1)
		}

		return pair
	}
}

func inputVarMapper(inputs []string) func(input string, _ int) string {
	return func(input string, _ int) string {
		if strings.Contains(input, varTasksModT) {
			return strings.ReplaceAll(input, varTasksModT, taskNameFilterJoin(tasks, func(t *runner.Task, _ int) bool { return t.Module }))
		}

		if strings.Contains(input, varTasksModF) {
			return strings.ReplaceAll(input, varTasksModF, taskNameFilterJoin(tasks, func(t *runner.Task, _ int) bool { return !t.Module }))
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
