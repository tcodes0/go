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
	"github.com/tcodes0/go/logging"
	"gopkg.in/yaml.v3"
)

const (
	taskModule int = iota + 1
	taskRepo

	needGitClean = "git-clean"
	needOnline   = "online"
)

type task struct {
	Env    map[string]string `yaml:"env"`
	Name   string            `yaml:"name"`
	Needs  string            `yaml:"needs"`
	Exec   []string          `yaml:"exec"`
	Kind   int               `yaml:"kind"`
	Inputs int               `yaml:"inputs"`
}

var (
	//go:embed config.yml
	config  string
	flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	tasks   []*task
)

func main() {
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs.")
	fColor := flagset.Bool("color", false, "colored logging output. (default false)")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usage(err)
		os.Exit(1)
	}

	err = yaml.Unmarshal([]byte(config), &tasks)
	if err != nil {
		usage(err)
		os.Exit(1)
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	err = run(*logger, os.Args[1:]...)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usage(err error) {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println()
	}

	fmt.Println("miscellaneous automation tool")
	fmt.Println("./run <task> \ttask args if any...")
	fmt.Println()
	fmt.Println("tasks available:")

	for _, task := range tasks {
		line := "./run "
		line += task.Name + "\t"

		if task.Kind == taskModule {
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

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}
}

func didYouMean(input string) {}

// run <task> <module or arg1> ...args.
func run(logger logging.Logger, args ...string) error {
	if len(args) == 0 {
		usage(nil)
		logger.Fatal(errors.New("task is required"))
	}

	task, found := lo.Find(tasks, func(t *task) bool { return t.Name == args[0] })
	if !found {
		usage(nil)
		didYouMean(args[0])
		logger.Fatal(fmt.Errorf("%s: unknown task", args[0]))
	}

	if task.Kind == taskModule {
		//nolint:mnd // len check
		if len(args) < 2 {
			usage(nil)
			logger.Fatal(fmt.Errorf("%s: module is required", task.Name))
		}

		mods, err := cmd.FindModules(logger)
		if err != nil {
			logger.Fatalf("FindModules: %v", err)
		}

		_, found := lo.Find(mods, func(m string) bool { return m == args[1] })
		if !found {
			usage(nil)
			didYouMean(args[1])
			logger.Fatal(fmt.Errorf("%s: invalid module", args[1]))
		}

		return moduleTask(logger, task, args[1], args[2:]...)
	}

	return repoTask(logger, task, args[1:]...)
}

func moduleTask(logger logging.Logger, task *task, module string, args ...string) error {
	return nil
}

func repoTask(logger logging.Logger, task *task, args ...string) error {
	return nil
}
