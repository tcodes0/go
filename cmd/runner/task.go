// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package runner

import (
	"errors"
	"fmt"
	"net/http"
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
	needGitClean = "<git-clean>"
	needOnline   = "<online>"
)

var ErrUsage = errors.New("see usage")

type Task struct {
	Env    []string `yaml:"env"`
	Name   string   `yaml:"name"`
	Needs  string   `yaml:"needs"`
	Exec   []string `yaml:"exec"`
	Module bool     `yaml:"module"`
}

// validate <module or input1> ...inputs.
func (task *Task) Validate(logger *logging.Logger, inputs ...string) error {
	_, help := lo.Find(inputs, func(input string) bool { return input == "-h" || input == "--help" })
	if help {
		return nil
	}

	err := task.ValidateModule(logger, inputs...)
	if err != nil {
		return err
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

func checkGitClean() error {
	err := exec.Command("git", "diff", "--exit-code").Run()
	if err != nil {
		return misc.Wrap(err, "please commit or stash all changes")
	}

	return nil
}

func checkOnline() error {
	//nolint:noctx // simple internet test
	res, err := http.Get("https://1.1.1.1" /*cloudflare*/)
	if err != nil {
		return misc.Wrap(err, "please check your internet connection")
	}

	defer res.Body.Close()

	return nil
}

func (task *Task) ValidateModule(logger *logging.Logger, inputs ...string) error {
	if !task.Module {
		return nil
	}

	if len(inputs) < 1 {
		return misc.Wrapf(ErrUsage, "%s: module is required", task.Name)
	}

	mods, err := cmd.FindModules(logger)
	if err != nil {
		return misc.Wrap(err, "FindModules")
	}

	_, found := lo.Find(mods, func(m string) bool { return m == inputs[0] })
	if !found {
		meant, ok := DidYouMean(inputs[0], mods)
		if ok {
			return misc.Wrapf(ErrUsage, "%s: unknown module, %s", inputs[0], meant)
		}

		return misc.Wrapf(ErrUsage, "%s: unknown module", inputs[0])
	}

	return nil
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
			// slice the last or first character until we match
			if i%2 == 0 {
				input = input[:len(input)-1]
			} else {
				input = input[1:]
			}
		}

		if len(matches) > limit {
			matches = matches[:limit]
		}
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
