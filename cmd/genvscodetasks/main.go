// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
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
	"github.com/tcodes0/go/jsonutil"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	VscodeRoot   string `yaml:"vscodeRoot"`
	TasksFile    string `yaml:"tasksFile"`
	ExtraModules string `yaml:"extraModules"`
}

var (
	//go:embed config.yml
	raw     string
	configs config
	flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
)

func main() {
	//nolint:lll // help string
	fModCmds := flagset.String("module-commands", "", "module commands to be added to the tasks.json file. Comma separated list, trimmed. Required.")
	//nolint:lll // help string
	fRepoCmds := flagset.String("repo-commands", "", "repo commands to be added to the tasks.json file. Comma separated list, trimmed. Required.")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	err = yaml.Unmarshal([]byte(raw), &configs)
	if err != nil {
		usageExit(err)
	}

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	if *fModCmds == "" || *fRepoCmds == "" {
		usageExit(errors.New("module-commands and repo-commands are required"))
	}

	err = generateVscodeTasks(*logger, *fModCmds, *fRepoCmds)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("generates vscode tasks.json file")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func generateVscodeTasks(logger logging.Logger, modInput, repoInput string) error {
	modCmds := lo.Map(strings.Split(modInput, ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	repoCmds := lo.Map(strings.Split(repoInput, ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	modules, err := cmd.FindModules(logger)
	if err != nil {
		return misc.Wrap(err, "findModules")
	}

	modules = slices.Concat(strings.Split(configs.ExtraModules, ","), modules)

	slices.Sort(modCmds)
	slices.Sort(repoCmds)

	filePath := configs.VscodeRoot + "/" + configs.TasksFile

	taskFile, err := readFile(filePath)
	if err != nil {
		return misc.Wrap(err, "readFile")
	}

	taskFile.Inputs[0].Options = modules
	taskFile.Inputs[1].Options = modCmds
	taskFile.Inputs[2].Options = repoCmds

	data := &bytes.Buffer{}
	encoder := json.NewEncoder(data)
	encoder.SetEscapeHTML(false)

	err = encoder.Encode(taskFile)
	if err != nil {
		return misc.Wrap(err, "encoding")
	}

	err = cmd.WriteFile(filePath, data.Bytes())
	if err != nil {
		return misc.Wrap(err, "writeFile")
	}

	command := exec.Command("prettier", "--write", filePath)

	err = command.Run()
	if err != nil {
		return misc.Wrapf(err, "formatting %s", command.Stderr)
	}

	return nil
}

type taskFile struct {
	Version string   `json:"version"`
	Tasks   []*task  `json:"tasks"`
	Inputs  []*input `json:"inputs"`
}

type task struct {
	Type    string   `json:"type"`
	Label   string   `json:"label"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Options struct {
		Cwd string `json:"cwd"`
	} `json:"options"`
	Presentation struct {
		Panel string `json:"panel"`
	} `json:"presentation"`
	ProblemMatcher []interface{} `json:"problemMatcher"`
}

type input struct {
	Type        string   `json:"type"`
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Default     string   `json:"default"`
	Options     []string `json:"options"`
}

func readFile(path string) (*taskFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, misc.Wrap(err, "opening")
	}

	defer file.Close()

	return jsonutil.UnmarshalReader[taskFile](file)
}
