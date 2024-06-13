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
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

const (
	needGitClean = "<git-clean>"
	needOnline   = "<online>"

	varModule    = "<module>"
	varCopy      = "<inherit>"
	varTasksModT = "<task-module-names>"
	varTasksModF = "<task-not-module-names>"

	envColor    = "CMD_COLOR"
	envLogLevel = "CMD_LOGLEVEL"
)

var (
	//go:embed config.yml
	config   string
	tasks    []*task
	errUsage = errors.New("see usage")
)

type task struct {
	Env    []string `yaml:"env"`
	Name   string   `yaml:"name"`
	Needs  string   `yaml:"needs"`
	Exec   []string `yaml:"exec"`
	Module bool     `yaml:"module"`
	Inputs int      `yaml:"inputs"`
}

// validate <module or arg1> ...args.
func (task *task) validate(logger logging.Logger, args ...string) error {
	err := task.validateModule(logger, args...)
	if err != nil {
		return err
	}

	if task.Inputs != 0 && task.Inputs != len(args) {
		return fmt.Errorf("%s: expected %d arguments got: %v", task.Name, task.Inputs, args)
	}

	for _, need := range strings.Split(task.Needs, ",") {
		need = strings.TrimSpace(need)

		switch need {
		default:
			err = fmt.Errorf("unknown need: %s", need)
		case "":
			continue
		case needGitClean:
			err = checkGitClean()
		case needOnline:
			err = checkOnline()
		}

		if err != nil {
			break
		}
	}

	return err
}

func (task *task) validateModule(logger logging.Logger, args ...string) error {
	if !task.Module {
		return nil
	}

	if len(args) < 1 {
		return misc.Wrapf(errUsage, "%s: module is required", task.Name)
	}

	mods, err := cmd.FindModules(logger)
	if err != nil {
		return misc.Wrap(err, "FindModules")
	}

	_, found := lo.Find(mods, func(m string) bool { return m == args[0] })
	if !found {
		didYouMean(args[0])

		return misc.Wrapf(errUsage, "%s: unknown module", args[0])
	}

	return nil
}

func main() {
	err := yaml.Unmarshal([]byte(config), &tasks)
	if err != nil {
		usage(err)
		os.Exit(1)
	}

	fColor := misc.LookupEnv(envColor, false)
	fLogLevel := misc.LookupEnv(envLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	err = run(*logger, os.Args[1:]...)
	if err != nil {
		if errors.Is(err, errUsage) {
			usage(err)
		}

		logger.Fatalf(err.Error())
	}
}

func usage(err error) {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println()
	}

	fmt.Println("miscellaneous automation tool")
	fmt.Println("usage: ./run <task> \ttask args if any...")
	fmt.Println()
	fmt.Println("tasks available:")

	for _, task := range tasks {
		line := "./run "
		line += task.Name + "\t"

		if task.Module {
			line += "<module>"
		} else {
			for range task.Inputs {
				line += "<arg>\t"
			}
		}

		fmt.Println(line)
	}

	modules, e := cmd.FindModules(logging.Logger{})
	if e != nil {
		fmt.Printf("finding modules: error: %v\n", e)
	}

	fmt.Println()
	fmt.Println("modules:", strings.Join(modules, ", "))
	fmt.Println("use 'all' as module to iterate all modules")
	fmt.Println("pass -h to commands to see further options")
	fmt.Println()
}

// run <task> <module or arg1> ...args.
func run(logger logging.Logger, args ...string) error {
	if len(args) == 0 {
		return misc.Wrap(errUsage, "task is required")
	}

	theTask, found := lo.Find(tasks, func(t *task) bool { return t.Name == args[0] })
	if !found {
		didYouMean(args[0])

		return misc.Wrapf(errUsage, "%s: unknown task", args[0])
	}

	err := theTask.validate(logger, args[1:]...)
	if err != nil {
		return err
	}

	for _, line := range theTask.Exec {
		cmdInput := slices.Concat(strings.Split(line, " "), args[1:])
		cmdInput = lo.Map(cmdInput, inputVarMapper)

		//nolint:gosec // has validation
		command := exec.Command(cmdInput[0], cmdInput[1:]...)
		logger.Debug().Log(strings.Join(cmdInput, " "))

		if len(theTask.Env) > 0 {
			envs := lo.Map(theTask.Env, envVarMapper(logger, args))
			command.Env = append(command.Env, envs...)
		}

		logger.Debug().Log("env: " + strings.Join(command.Env, " "))

		out, err := command.Output()
		if err != nil {
			exitErr, ok := (err).(*exec.ExitError)
			if ok {
				logger.Error().Log("stderr: " + string(exitErr.Stderr))
			}

			return misc.Wrapf(err, "command '%s'", strings.Join(cmdInput, " "))
		}

		if len(out) > 0 {
			logger.Log("\n" + string(out))
		}
	}

	return nil
}

func didYouMean(input string) {}

func checkGitClean() error {
	err := exec.Command("git", "diff", "--exit-code").Run()
	if err != nil {
		return misc.Wrap(err, "please commit or stash all changes")
	}

	return nil
}

func checkOnline() error {
	//nolint:noctx // simple internet test
	res, err := http.Get("1.1.1.1" /*cloudflare*/)
	if err != nil {
		return misc.Wrap(err, "please check your internet connection")
	}

	defer res.Body.Close()

	return nil
}

func envVarMapper(logger logging.Logger, args []string) func(pair string, _ int) string {
	return func(pair string, _ int) string {
		out := strings.Replace(pair, varModule, args[1], 1)

		if strings.Contains(out, varCopy) {
			key := strings.Split(out, "=")[0]

			val, ok := os.LookupEnv(key)
			if !ok {
				logger.Warn().Logf("env value inherited is empty: " + key)
			}

			out = strings.Replace(out, varCopy, val, 1)
		}

		return out
	}
}

func inputVarMapper(input string, _ int) string {
	switch input {
	default:
		return input
	case varTasksModT:
		return taskNameFilterJoin(tasks, func(t *task, _ int) bool { return t.Module })
	case varTasksModF:
		return taskNameFilterJoin(tasks, func(t *task, _ int) bool { return !t.Module })
	}
}

func taskNameFilterJoin(tasks []*task, filterFunc func(t *task, _ int) bool) string {
	modTasks := lo.Filter(tasks, filterFunc)
	names := lo.Reduce(modTasks, func(agg []string, t *task, _ int) []string {
		return append(agg, t.Name)
	}, make([]string, 0, len(modTasks)))

	return strings.Join(names, ",")
}
