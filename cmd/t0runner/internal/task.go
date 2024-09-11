// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/hue"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

const (
	varPackage = "<package>" // the package passed as input
	varInherit = "<inherit>" // copy this from the environment
	varSpace   = "<space>"   // single whitespace
)

var ErrUsage = errors.New("see usage")

type Task struct {
	Name        string `yaml:"name"`
	PackageName string
	Env         []string `yaml:"env"`
	Exec        []string `yaml:"exec"`
	inputs      []string
	Package     bool `yaml:"package"`
}

// input[0] is task name, input[1] is package name.
func (task *Task) SetInputs(inputs []string) {
	if task.Package {
		if strings.HasSuffix(inputs[1], "/") {
			// remove trailing slash, allows TAB to complete valid packages
			inputs[1] = inputs[1][:len(inputs[1])-1]
		}

		task.PackageName = inputs[1]
		task.inputs = inputs[2:]

		return
	}

	task.inputs = inputs[1:]

	return
}

func (task *Task) validatePackage(logger *logging.Logger) error {
	if !task.Package {
		return nil
	}

	_, help := lo.Find(task.inputs, func(input string) bool { return input == "-h" || input == "--help" })
	if help {
		return nil
	}

	if task.PackageName == "" {
		return misc.Wrapf(ErrUsage, "%s: package is required", task.Name)
	}

	pkgs, err := cmd.FindPackages(logger)
	if err != nil {
		return misc.Wrapfl(err)
	}

	_, found := lo.Find(pkgs, func(m string) bool { return m == task.PackageName })
	if !found {
		meant, ok := DidYouMean(task.PackageName, pkgs)
		if ok {
			return misc.Wrapf(ErrUsage, "%s: unknown package, %s", task.PackageName, meant)
		}

		return misc.Wrapf(ErrUsage, "%s: unknown package", task.PackageName)
	}

	return nil
}

func (task *Task) Execute(logger *logging.Logger) error {
	err := task.validatePackage(logger)
	if err != nil {
		return misc.Wrapfl(err)
	}

	for _, line := range task.Exec {
		cmdInput := slices.Concat(strings.Split(line, " "), task.inputs)
		cmdInput = lo.Map(cmdInput, varMapper(task))
		cmdInput = lo.Map(cmdInput, unescapeMapper)

		logger.Debug(cmdInput)

		//nolint:gosec // has validation
		command := exec.Command(cmdInput[0], cmdInput[1:]...)

		logger.Debug(line)

		if len(task.Env) > 0 {
			envs := lo.Map(task.Env, envVarMapper(task, logger))
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
			logger.Info("command stderr")
			fmt.Fprint(os.Stderr, hue.TermColor(hue.Gray)+stderrBuffer.String()+hue.End)
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

func envVarMapper(task *Task, logger *logging.Logger) func(pair string, _ int) string {
	return func(pair string, _ int) string {
		if strings.Contains(pair, varPackage) {
			return strings.Replace(pair, varPackage, task.PackageName, 1)
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

func varMapper(task *Task) func(input string, _ int) string {
	return func(input string, _ int) string {
		if strings.Contains(input, varPackage) {
			return strings.ReplaceAll(input, varPackage, task.PackageName)
		}

		if strings.Contains(input, varSpace) {
			return strings.ReplaceAll(input, varSpace, " ")
		}

		return input
	}
}

func unescapeMapper(input string, _ int) string {
	// literal # is desired but is considered yaml comment
	return strings.ReplaceAll(input, `\#`, "#")
}

func DidYouMean(input string, candidates []string) (string, bool) {
	type match struct {
		word  string
		score int
	}

	input = strings.ToLower(input)
	limit := 5
	matches := make([]match, 0, len(candidates))

	for i := range len(input) {
		matches = lo.FilterMap(candidates, func(w string, _ int) (match, bool) {
			m := match{word: w, score: fuzzy.RankMatch(input, w)}
			if m.score == -1 {
				return m, false
			}

			return m, true
		})

		if len(matches) == 0 {
			// on even iterations slice the last
			// on odd slice the first character; until we match
			if i%2 == 0 {
				input = input[:len(input)-1]
			} else {
				input = input[1:]
			}

			continue
		}

		if len(matches) > limit {
			matches = matches[:limit]
		}

		break
	}

	if len(matches) == 0 {
		return "", false
	}

	slices.SortFunc(matches, func(a, b match) int {
		return b.score - a.score
	})

	return lo.Reduce(matches, func(agg string, item match, i int) string {
		if i == len(matches)-1 {
			return agg + "'" + item.word + "'?"
		}

		return agg + "'" + item.word + "'" + ", "
	}, "did you mean: "), true
}
