// Copyright 2024 Raphael Thomazella. All rights reserved.
//  Use of this source code is governed by the BSD-3-Clause
//  license that can be found in the LICENSE file and online
//  at https://opensource.org/license/BSD-3-clause.

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
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/samber/lo"
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
	//nolint:gosec // not a real token
	varToken = "<github-token>"
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

	err = ci(*logger, ciEvent)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}

	logger.Logf("Took %d", time.Since(start)/time.Second)
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

func ci(logger logging.Logger, theEvent *event) error {
	eventJSONFile, ciLogFile, gitBranch, token, err := prepare(logger, theEvent)
	if err != nil {
		return err
	}

	eventVarResolver(theEvent, gitBranch)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmds := cmdVarResolver(configs.Exec, eventJSONFile, token)
	//nolint:gosec // has validation
	ciCmd := exec.Command(cmds[0], strings.Join(cmds[1:], " "))
	ciCmd.Stdout = outBuf
	ciCmd.Stderr = errBuf
	renderDone := make(chan bool)

	outBuf.WriteString("stdout:\n")
	errBuf.WriteString("\nstderr:\n")

	defer func() {
		flushLogs(logger, ciLogFile, outBuf, errBuf)
		renderDone <- true
	}()

	err = ciCmd.Start()
	if err != nil {
		return misc.Wrap(err, "start")
	}

	tick := time.NewTicker(misc.Seconds(2))

	go func() {
		select {
		case <-tick.C:
			render(outBuf, errBuf)
		case <-renderDone:
			return
		}
	}()

	errChan := make(chan error, 1)

	go func() {
		errChan <- misc.Wrap(ciCmd.Wait(), "wait")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
			logger.Debug().Log("\ncaught %s", sig.String())
			errChan <- errors.New(sig.String())
		}
	}()

	return <-errChan
}

func prepare(logger logging.Logger, theEvent *event) (eventFile, ciLog, branch, token string, err error) {
	command := exec.Command("mktemp", "/tmp/ci-event-json-XXXXXX")

	eventJSONFile, err := command.Output()
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "mktemp event json")
	}

	logger.Debug().Log("event json file %s", eventJSONFile)

	data, err := json.Marshal(theEvent)
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "marshal")
	}

	command = exec.Command("mktemp", "/tmp/ci-log-json-XXXXXX")

	ciLogFile, err := command.Output()
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "mktemp ci log")
	}

	logger.Debug().Log("ci log file %s", ciLogFile)

	err = cmd.WriteFile(string(eventJSONFile), data)
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "writeFile event")
	}

	err = cmd.WriteFile(string(ciLogFile), data)
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "writeFile ci")
	}

	command = exec.Command("git", "branch", "--show-current")

	gitBranch, err := command.Output()
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "git branch")
	}

	command = exec.Command("gh", "auth", "token")

	bToken, err := command.Output()
	if err != nil {
		return "", "", "", "", misc.Wrap(err, "auth token")
	}

	logger.Debug().Log("token length %d", len(token))

	return string(eventJSONFile), string(ciLogFile), string(gitBranch), string(bToken), nil
}

func eventVarResolver(theEvent *event, gitBranch string) {
	if theEvent.PullRequest != nil {
		if strings.Contains(theEvent.PullRequest.Base.Ref, varLocalBranch) {
			theEvent.PullRequest.Base.Ref = strings.Replace(theEvent.PullRequest.Base.Ref, varLocalBranch, gitBranch, 1)
		}

		if strings.Contains(theEvent.PullRequest.Head.Ref, varLocalBranch) {
			theEvent.PullRequest.Head.Ref = strings.Replace(theEvent.PullRequest.Head.Ref, varLocalBranch, gitBranch, 1)
		}
	}

	if theEvent.Push != nil {
		if strings.Contains(theEvent.Push.BaseRef, varLocalBranch) {
			theEvent.Push.BaseRef = strings.Replace(theEvent.Push.BaseRef, varLocalBranch, gitBranch, 1)
		}
	}
}

func flushLogs(logger logging.Logger, logfile string, stdout, stderr *bytes.Buffer) {
	_, err := stdout.Write(stderr.Bytes())
	if err != nil {
		logger.Error().Logf("concat buffers: %s", err.Error())

		return
	}

	err = cmd.WriteFile(logfile, stdout.Bytes())
	if err != nil {
		logger.Error().Logf("writeFile: %s", err.Error())
	}
}

func cmdVarResolver(inputStr, eventJSONFile, token string) []string {
	return lo.Map(strings.Split(inputStr, " "), func(input string, _ int) string {
		if strings.Contains(input, varEventJSONFile) {
			return strings.Replace(input, varEventJSONFile, eventJSONFile, 1)
		}

		if strings.Contains(input, varToken) {
			return strings.Replace(input, varToken, token, 1)
		}

		return input
	})
}

func render(stdout, stderr *bytes.Buffer) {}
