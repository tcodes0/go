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
	"strings"

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
func (task *Task) Validate(logger logging.Logger, inputs ...string) error {
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

func (task *Task) ValidateModule(logger logging.Logger, inputs ...string) error {
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
		DidYouMean(inputs[0])

		return misc.Wrapf(ErrUsage, "%s: unknown module", inputs[0])
	}

	return nil
}

func DidYouMean(input string) {}
