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
	Env     []string `yaml:"env"`
	Name    string   `yaml:"name"`
	Exec    []string `yaml:"exec"`
	Package bool     `yaml:"package"`
}

// validate <package or input1> ...inputs.
func (task *Task) Validate(logger *logging.Logger, inputs ...string) error {
	_, help := lo.Find(inputs, func(input string) bool { return input == "-h" || input == "--help" })
	if help {
		return nil
	}

	err := task.ValidatePackage(logger, inputs...)
	if err != nil {
		return err
	}

	return err
}

func (task *Task) ValidatePackage(logger *logging.Logger, inputs ...string) error {
	if !task.Package {
		return nil
	}

	if len(inputs) < 1 {
		return misc.Wrapf(ErrUsage, "%s: package is required", task.Name)
	}

	mods, err := cmd.FindPackages(logger)
	if err != nil {
		return misc.Wrap(err, "FindPackages")
	}

	_, found := lo.Find(mods, func(m string) bool { return m == inputs[0] })
	if !found {
		meant, ok := DidYouMean(inputs[0], mods)
		if ok {
			return misc.Wrapf(ErrUsage, "%s: unknown package, %s", inputs[0], meant)
		}

		return misc.Wrapf(ErrUsage, "%s: unknown package", inputs[0])
	}

	return nil
}

func (task *Task) Execute(logger *logging.Logger, tasks []*Task, inputs ...string) error {
	for _, line := range task.Exec {
		cmdInput := slices.Concat(strings.Split(line, " "), inputs[1:])
		cmdInput = lo.Map(cmdInput, varMapper(inputs))
		cmdInput = lo.Map(cmdInput, unescapeMapper)

		//nolint:gosec // has validation
		command := exec.Command(cmdInput[0], cmdInput[1:]...)

		logger.Debug(line)

		if len(task.Env) > 0 {
			envs := lo.Map(task.Env, envVarMapper(inputs, logger))
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
			logger.Info("command stderr BEGINS")
			fmt.Fprint(os.Stderr, stderrBuffer.String())
			logger.Info("command stderr ENDS")
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

func envVarMapper(inputs []string, logger *logging.Logger) func(pair string, _ int) string {
	return func(pair string, _ int) string {
		if strings.Contains(pair, varPackage) {
			return strings.Replace(pair, varPackage, inputs[1], 1)
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

func varMapper(inputs []string) func(input string, _ int) string {
	return func(input string, _ int) string {
		if strings.Contains(input, varPackage) {
			return strings.ReplaceAll(input, varPackage, inputs[1])
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