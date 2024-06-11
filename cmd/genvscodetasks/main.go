// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/jsonutil"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var (
	flagset      = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	vscoderoot   = ".vscode"
	tasksFile    = "tasks.json"
	extraModules = []string{"all"}
	ignore       = regexp.MustCompile(`test$|\.local.*|cmd/template`)
)

func main() {
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs.")
	fColor := flagset.Bool("color", false, "colored logging output. (default false)")
	//nolint:lll // long string
	fModCmds := flagset.String("module-commands", "", "module commands to be added to the tasks.json file. Comma separated list, trimmed. Required.")
	//nolint:lll // long string
	fRepoCmds := flagset.String("repo-commands", "", "repo commands to be added to the tasks.json file. Comma separated list, trimmed. Required.")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
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

	modules, err := findModules(logger)
	if err != nil {
		return misc.Wrap(err, "findModules")
	}

	slices.Sort(modCmds)
	slices.Sort(repoCmds)

	filePath := vscoderoot + "/" + tasksFile

	taskFile, err := readFile(filePath)
	if err != nil {
		return misc.Wrap(err, "readFile")
	}

	taskFile.Inputs[0].Options = modules
	taskFile.Inputs[1].Options = modCmds
	taskFile.Inputs[2].Options = repoCmds

	err = writeFile(filePath, taskFile)
	if err != nil {
		return misc.Wrap(err, "writeFile")
	}

	cmd := exec.Command("prettier", "--write", filePath)

	err = cmd.Run()
	if err != nil {
		return misc.Wrapf(err, "formatting %s", cmd.Stderr)
	}

	return nil
}

func findModules(logger logging.Logger) ([]string, error) {
	cmd := exec.Command("find", ".", "-mindepth", "2", "-maxdepth", "3", "-type", "f", "-name", "*.go", "-exec", "dirname", "{}", ";")

	findOut, err := cmd.CombinedOutput()
	if err != nil {
		return nil, misc.Wrapf(err, "finding, %s", findOut)
	}

	logger.Debug().Logf("find output: %s", findOut)

	modules := strings.Split(string(findOut), "\n")
	modules = slices.Concat(modules, extraModules)
	modules = lo.Uniq(modules)

	out := make([]string, 0, len(modules))

	for _, module := range modules {
		if ignore.MatchString(module) {
			logger.Debug().Logf("ignored %s", module)

			continue
		}

		if module == "" {
			continue
		}

		out = append(out, strings.Replace(module, "./", "", 1))
	}

	slices.Sort(out)

	return out, nil
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

func writeFile(path string, taskFile *taskFile) error {
	data := &bytes.Buffer{}
	encoder := json.NewEncoder(data)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(taskFile)
	if err != nil {
		return misc.Wrap(err, "encoding")
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return misc.Wrap(err, "opening")
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return misc.Wrap(err, "stat")
	}

	if int64(data.Len()) < stat.Size() {
		// new file is smaller, truncate to new size
		err = file.Truncate(int64(data.Len()))
		if err != nil {
			return misc.Wrap(err, "truncating")
		}
	}

	_, err = file.Write(data.Bytes())
	if err != nil {
		return misc.Wrap(err, "writing")
	}

	return nil
}
