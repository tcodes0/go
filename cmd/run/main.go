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

type taskType int

const (
	taskTypeModule taskType = iota + 1
	taskTypeRepo
)

type task struct {
	Name   string   `yaml:"name"`
	Kind   taskType `yaml:"kind"`
	Inputs int      `yaml:"inputs"`
}

var (
	//go:embed config.yml
	config         string
	flagset        = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	errUnknownTask = errors.New("unknown task")
	tasks          []*task
)

func main() {
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs.")
	fColor := flagset.Bool("color", false, "colored logging output. (default false)")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	err = yaml.Unmarshal([]byte(config), &tasks)
	if err != nil {
		usageExit(err)
	}

	var fTask string

	if len(os.Args) > 1 {
		fTask = os.Args[1]
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	if fTask == "" {
		usageExit(errors.New("task is required"))
	}

	task, found := lo.Find(tasks, func(t *task) bool { return t.Name == fTask })
	if !found {
		usageExit(errUnknownTask)
	}

	err = run(*logger, task)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("miscellaneous automation tool")
	fmt.Println("./run <task> \ttask args if any...")
	fmt.Println()
	fmt.Println("tasks available:")

	for _, task := range tasks {
		line := "./run "
		line += task.Name + "\t"

		if task.Kind == taskTypeModule {
			line += "<module>"
		} else {
			for i := range task.Inputs {
				line += fmt.Sprintf("<arg %d> ", i+1)
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

	if errors.Is(err, errUnknownTask) {
		didYouMean(os.Args[1])
	}

	os.Exit(1)
}

func didYouMean(input string) {}

func run(_ logging.Logger, task *task) error {
	return nil
}
