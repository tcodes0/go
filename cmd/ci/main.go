// Copyright 2024 Raphael Thomazella. All rights reserved.
//  Use of this source code is governed by the BSD-3-Clause
//  license that can be found in the LICENSE file and online
//  at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type config struct {
	PassFailRegexp *regexp.Regexp
	PRTitle        string `yaml:"prTitle"`
	PRBaseRef      string `yaml:"prBaseRef"`
	PRHeadRef      string `yaml:"prHeadRef"`
	PushBaseRef    string `yaml:"pushBaseRef"`
	Exec           string `yaml:"exec"`
	MinLines       int    `yaml:"minLines"`
	MinDurationRaw int    `yaml:"minDurationSeconds"`
	MaxDurationRaw int    `yaml:"maxDurationSeconds"`
	MinDuration    time.Duration
	MaxDuration    time.Duration
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
	//go:embed config.yml
	raw     string
	configs config
	logger  = &logging.Logger{}
	flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	varToken         = "<github-token>"
	varLocalBranch   = "<local-branch>"
	varEventJSONFile = "<event-json-file>"

	passFailRegexpRaw = `\[(?P<job>[^]]+)\].*(?P<status>succeeded|failed)`
	pass              = "succeeded"
)

func main() {
	start := time.Now()

	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Fatalf("%v", msg)
		}

		logger.Logf("took %ds", time.Since(start)/time.Second)

		if err != nil {
			logger.Error().Log(err.Error())
			os.Exit(1)
		}
	}()

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	fPush := flagset.Bool("push", false, "use a push event, what happens on merge")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	err = yaml.Unmarshal([]byte(raw), &configs)
	if err != nil {
		usageExit(err)
	}

	configs.PassFailRegexp = regexp.MustCompile(passFailRegexpRaw)
	configs.MinDuration = misc.Seconds(configs.MinDurationRaw)
	configs.MaxDuration = misc.Seconds(configs.MaxDurationRaw)

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
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("runs ci locally")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		logger.Error().Log(err.Error())
	}

	os.Exit(1)
}

func ci(logger logging.Logger, theEvent *event) error {
	eventJSONFile, ciLogFile, token, err := prepareEnv(logger, theEvent)
	if err != nil {
		return err
	}

	ctx, ciCmd, ciStdout, ciStderr, cancel := prepareCI(eventJSONFile, token)
	defer cancel()

	err = ciCmd.Start()
	if err != nil {
		return misc.Wrap(err, "start")
	}

	ticker := time.NewTicker(misc.Seconds(1))
	defer ticker.Stop()

	requestFrame := make(chan bool)
	go render(ctx, ciStdout, ciStderr, ticker, requestFrame)

	go misc.RoutineListenStopSignal(ctx, func(sig os.Signal) {
		logger.Error().Log("\nfatal: %s stopping...", sig.String())
		// clear cursor to end of screen
		fmt.Printf("\033[0J")

		e := ciCmd.Process.Signal(sig)
		if e != nil {
			logger.Error().Logf("ci %s: %s", sig.String(), e.Error())
		}
	})

	defer func() {
		flushLogs(logger, ciLogFile, ciStdout, ciStderr)
		// ensure all info is on screen
		requestFrame <- true
	}()

	return misc.Wrap(ciCmd.Wait(), "wait")
}

//nolint:funlen // it does a lot!
func prepareEnv(logger logging.Logger, theEvent *event) (eventFile, ciLogFile, token string, e error) {
	errG := errgroup.Group{}
	errG.Go(func() error {
		eventFileBytes, err := exec.Command("mktemp", "/tmp/ci-event-json-XXXXXX").Output()
		eventFile = strings.TrimSuffix(string(eventFileBytes), "\n")
		logger.Debug().Logf("event json file %s", eventFile)

		return misc.Wrap(err, "mktemp event json")
	})

	errG.Go(func() error {
		bGitBranch, err := exec.Command("git", "branch", "--show-current").Output()
		gitBranch := strings.TrimSuffix(string(bGitBranch), "\n")
		eventVarResolver(theEvent, gitBranch)
		logger.Debug().Logf("git branch %s", gitBranch)

		return misc.Wrap(err, "git branch")
	})

	errG.Go(func() error {
		ciLogBytes, err := exec.Command("mktemp", "/tmp/ci-log-json-XXXXXX").Output()
		ciLogFile = strings.TrimSuffix(string(ciLogBytes), "\n")
		logger.Debug().Logf("ci log file %s", ciLogFile)

		return misc.Wrap(err, "mktemp ci log")
	})

	errG.Go(func() error {
		bToken, err := exec.Command("gh", "auth", "token").Output()
		token = strings.TrimSuffix(string(bToken), "\n")
		logger.Debug().Logf("token length %d", len(token))

		return misc.Wrap(err, "auth token")
	})

	e = errG.Wait()
	if e != nil {
		return "", "", "", e
	}

	data, err := json.Marshal(theEvent)
	if err != nil {
		return "", "", "", misc.Wrap(err, "marshal")
	}

	errG.Go(func() error {
		err := cmd.WriteFile(eventFile, data)

		return misc.Wrap(err, "writeFile event")
	})

	errG.Go(func() error {
		err := cmd.WriteFile(ciLogFile, data)

		return misc.Wrap(err, "writeFile ci")
	})

	e = errG.Wait()
	if e != nil {
		return "", "", "", e
	}

	return eventFile, ciLogFile, token, nil
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

func prepareCI(eventJSONFile, token string) (
	ctx context.Context,
	ciCmd *exec.Cmd,
	stdout,
	stderr *bytes.Buffer,
	cancel context.CancelFunc,
) {
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	cmds := cmdVarResolver(configs.Exec, eventJSONFile, token)
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(configs.MaxDuration))
	//nolint:gosec // has validation
	ciCmd = exec.CommandContext(ctx, cmds[0], cmds[1:]...)
	ciCmd.Stdout = stdout
	ciCmd.Stderr = stderr

	return ctx, ciCmd, stdout, stderr, cancel
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

func render(ctx context.Context, stdout, _ *bytes.Buffer, ticker *time.Ticker, request chan bool) {
	fmt.Print("\033[H\033[2J") // move 1-1, clear whole screen
	fmt.Printf("running ci...\n")

	for {
		select {
		case <-ctx.Done():
			return

		case <-request:
		case <-ticker.C:
			matches := configs.PassFailRegexp.FindAllStringSubmatch(stdout.String(), -1)

			fmt.Println("\033[H") // move 1-1

			for _, match := range matches {
				job := strings.Trim(match[1], " ")

				if match[2] == pass {
					fmt.Printf("%s %s%s%s\n", "\033[7;38;05;242m PASS \033[0m", "\033[2m", job, "\033[0m")
				} else {
					fmt.Printf("%s %s\n", "\033[2;7;38;05;197;47m FAIL \033[0m", job)
				}
			}
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
