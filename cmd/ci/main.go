// Copyright 2024 Raphael Thomazella. All rights reserved.
//  Use of this source code is governed by the BSD-3-Clause
//  license that can be found in the LICENSE file and online
//  at https://opensource.org/license/BSD-3-clause.

package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

type config struct {
	StatusRegexp *regexp.Regexp
	PRTitle      string        `yaml:"prTitle"`
	PRBaseRef    string        `yaml:"prBaseRef"`
	PRHeadRef    string        `yaml:"prHeadRef"`
	PushBaseRef  string        `yaml:"pushBaseRef"`
	StatusRaw    string        `yaml:"statusRegexp"`
	Exec         string        `yaml:"exec"`
	MinLines     int           `yaml:"minLines"`
	MinDuration  time.Duration `yaml:"minDurationSeconds"`
}

type event struct {
	//nolint:tagliatelle // vendor defined
	PullRequest *fieldPullRequest `json:"pull_request"`
	Push        *fieldPush        `json:"push"`
	Local       bool              `json:"local"`
}

type fieldPullRequest struct {
	Title string `json:"title"`
	Head  ref    `json:"head"`
	Base  ref    `json:"base"`
}

type fieldPush struct {
	//nolint:tagliatelle // vendor defined
	BaseRef string `json:"base_ref"`
}

type ref struct {
	Ref string `json:"ref"`
}

var (
	flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	//go:embed config.yml
	raw              string
	configs          config
	varLocalBranch   = "<local-branch>"
	varEventJSONFile = "<event-json-file>"
	varToken         = "<github-token>"
)

func main() {
	start := time.Now()
	fPush := flagset.Bool("push", false, "use a push event, what happens on merge")

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
	configs.StatusRegexp = regexp.MustCompile(configs.StatusRaw)

	ciEvent := &event{
		PullRequest: &fieldPullRequest{
			Title: configs.PRTitle,
			Head:  ref{Ref: configs.PRHeadRef},
			Base:  ref{Ref: configs.PRBaseRef},
		},
		Local: true,
	}

	if *fPush {
		ciEvent.Push = &fieldPush{BaseRef: configs.PushBaseRef}
		ciEvent.PullRequest = nil
	}

	err = ci(*logger, ciEvent, start)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("runs ci locally")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func ci(logger logging.Logger, theEvent *event, start time.Time) error {
	return nil
}
