// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package logging_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/logging"
	"golang.org/x/sync/errgroup"
)

var (
	debug = "DEBUG "
	info  = "INFO "
	warn  = "WARN "
	erro  = "ERROR "
	fatal = "FATAL "

	// sequences of control chars to end or start formatting
	// matches one sequence after the other.
	rawANSI = `[0-9;\033\[m]+`
	// date, time, file:line.
	rawFlags = `\d{4}/\d{2}/\d{2}` + " " + `\d{2}:\d{2}:\d{2}` + " " + `[a-z_]+\.go:\d+`

	matchEmpty = regexp.MustCompile("^$")
	matchDebug = regexp.MustCompile(debug + rawFlags + ": testing\n")
	matchInfo  = regexp.MustCompile(info + rawFlags + ": testing\n")
	matchWarn  = regexp.MustCompile(warn + rawFlags + ": testing\n")
	matchError = regexp.MustCompile(erro + rawFlags + ": testing\n")
	matchFatal = regexp.MustCompile(fatal + rawFlags + ": testing\n")
)

func TestLoggerRace(t *testing.T) {
	t.Parallel()

	logger := logging.Create(logging.OptWriter(io.Discard), logging.OptExit(func(int) {}))
	group := errgroup.Group{}

	group.Go(func() error {
		logger.Debug("routine 1")
		logger.Info("routine 1")
		logger.Warn("routine 1")
		logger.SetLevel(logging.LWarn)
		logger.Error("routine 1")
		logger.Fatal("routine 1")

		return nil
	})

	group.Go(func() error {
		logger.Debug("routine 2")
		logger.Info("routine 2")
		logger.Warn("routine 2")
		logger.Error("routine 2")
		logger.Fatal("routine 2")

		return nil
	})

	require.NoError(t, group.Wait(), "wait")
}

func TestLoggerOutput(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
	testCalls := [][]any{
		{"Debugf", "test%s", "ing"},
		{"Info", "testing"},
		{"Warn", "testing"},
		{"Errorf", "test%s", "ing"},
		{"Fatalf", "test%s", "ing"},
	}

	for callN, call := range testCalls {
		name := fmt.Sprintf("[%d] %s", callN, call[0])
		expected := []*regexp.Regexp{matchDebug, matchInfo, matchWarn, matchError, matchFatal}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptLevel(logging.LDebug))
			testRun(assert, name, logger, call, out, expected[callN])
		})
	}

	for callN, call := range testCalls {
		name := fmt.Sprintf("[%d] color %s", callN, call[0])
		expected := []*regexp.Regexp{
			regexp.MustCompile(rawANSI + debug + rawANSI + rawFlags + ": " + rawANSI + "testing\n"),
			regexp.MustCompile(rawANSI + info + rawFlags + ": " + rawANSI + "testing\n"),
			regexp.MustCompile(rawANSI + warn + rawANSI + rawFlags + ": " + rawANSI + "testing\n"),
			regexp.MustCompile(rawANSI + erro + rawANSI + rawFlags + ": " + rawANSI + "testing\n"),
			regexp.MustCompile(rawANSI + fatal + rawANSI + rawFlags + ": " + rawANSI + "testing\n"),
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptLevel(logging.LDebug), logging.OptColor())
			testRun(assert, name, logger, call, out, expected[callN])
		})
	}

	t.Run("noop logger", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		logger := &logging.Logger{}
		testRun(assert, "noop logger", logger, []any{"Fatal", "testing"}, out, matchEmpty)
	})
}

func testRun(
	assert *require.Assertions, name string, logger *logging.Logger, call []any, out *bytes.Buffer, reg *regexp.Regexp,
) {
	//nolint:forcetypeassert // test
	method := reflect.ValueOf(logger).MethodByName((call[0]).(string))
	if !method.IsValid() {
		assert.Fail("invalid method", name)
	}

	args := make([]reflect.Value, len(call)-1)
	for i, arg := range call[1:] {
		args[i] = reflect.ValueOf(arg)
	}

	ret := method.Call(args)
	assert.Empty(ret, name)

	if reg != nil {
		assert.Regexp(reg, out.String(), "expected output on '%s'", name)
	}
}

func TestLoggerData(t *testing.T) {
	t.Parallel()

	assert := require.New(t)
	mockMap := map[string]any{
		"hello": "world",
		"foo":   "bar",
	}

	t.Run("infoData", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}))
		testRun(assert, "infoData", logger, []any{"InfoData", mockMap, "testing"}, out, regexp.MustCompile(
			info+rawFlags+": "+`\(hello|foo=world|bar, hello|foo=world|bar\) `+"testing\n",
		))
	})

	t.Run("errorData color", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptColor())
		testRun(assert, "errorData color", logger, []any{"ErrorData", mockMap, "testing"}, out, regexp.MustCompile(
			rawANSI+erro+rawANSI+rawFlags+": "+rawANSI+"hello|foo"+rawANSI+"="+rawANSI+"world|bar"+
				rawANSI+", "+rawANSI+"foo|hello"+rawANSI+"="+rawANSI+"bar|world"+rawANSI+" testing",
		))
	})
}

func TestLoggerLevel(t *testing.T) {
	t.Parallel()

	testCalls := [][]any{
		{"Debug", "testing"},
		{"Info", "testing"},
		{"Warn", "testing"},
		{"Error", "testing"},
		{"Fatal", "testing"},
	}
	expected := [][]*regexp.Regexp{
		/* 1 Debug*/ {matchDebug, matchInfo, matchWarn, matchError, matchFatal},
		/* 2 Info*/ {matchEmpty, matchInfo, matchWarn, matchError, matchFatal},
		/* 3 Warn */ {matchEmpty, matchEmpty, matchWarn, matchError, matchFatal},
		/* 4 Error*/ {matchEmpty, matchEmpty, matchEmpty, matchError, matchFatal},
		/* 5 Fatal*/ {matchEmpty, matchEmpty, matchEmpty, matchEmpty, matchFatal},
		/* 6 None*/ {matchEmpty, matchEmpty, matchEmpty, matchEmpty, matchEmpty},
	}

	for i := range int(logging.LNone) {
		level := i + 1
		name := fmt.Sprintf("level %d", level)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for callN, call := range testCalls {
				name := fmt.Sprintf("call [%d] level %d", callN, level)
				out := &bytes.Buffer{}
				logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptLevel(logging.Level(level)))
				testRun(require.New(t), name, logger, call, out, expected[i][callN])
			}
		})
	}
}
