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
	ctrlSeqRE  = `[0-9;\033\[m]+`
	reDate     = `\d{4}/\d{2}/\d{2}`
	reTime     = `\d{2}:\d{2}:\d{2}`
	reFileLine = `[a-z_]+\.go:\d+`
	fullRE     = reDate + " " + reTime + " " + reFileLine

	outREDebug = regexp.MustCompile(debug + fullRE + ": testing\n")
	outREInfo  = regexp.MustCompile(info + fullRE + ": testing\n")
	outREWarn  = regexp.MustCompile(warn + fullRE + ": testing\n")
	outREError = regexp.MustCompile(erro + fullRE + ": testing\n")
	outREFatal = regexp.MustCompile(fatal + fullRE + ": testing\n")
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

//nolint:funlen // test
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
		expected := []*regexp.Regexp{outREDebug, outREInfo, outREWarn, outREError, outREFatal}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptLevel(logging.LDebug))
			testRun(assert, name, logger, call, out, expected[callN], nil)
		})
	}

	for callN, call := range testCalls {
		name := fmt.Sprintf("[%d] color %s", callN, call[0])
		expected := []*regexp.Regexp{
			regexp.MustCompile(ctrlSeqRE + debug + ctrlSeqRE + fullRE + ": " + ctrlSeqRE + "testing\n"),
			regexp.MustCompile(ctrlSeqRE + info + fullRE + ": " + ctrlSeqRE + "testing\n"),
			regexp.MustCompile(ctrlSeqRE + warn + ctrlSeqRE + fullRE + ": " + ctrlSeqRE + "testing\n"),
			regexp.MustCompile(ctrlSeqRE + erro + ctrlSeqRE + fullRE + ": " + ctrlSeqRE + "testing\n"),
			regexp.MustCompile(ctrlSeqRE + fatal + ctrlSeqRE + fullRE + ": " + ctrlSeqRE + "testing\n"),
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptLevel(logging.LDebug), logging.OptColor())
			testRun(assert, name, logger, call, out, expected[callN], nil)
		})
	}

	t.Run("noop logger", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		logger := &logging.Logger{}
		testRun(assert, "noop logger", logger, []any{"Info", "testing"}, out, regexp.MustCompile("^$"), nil)
	})
}

func testRun(
	assert *require.Assertions, name string, logger *logging.Logger, call []any, out *bytes.Buffer, reg, notReg *regexp.Regexp,
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

	if notReg != nil {
		assert.NotRegexp(notReg, out.String(), "unexpected output on '%s'", name)
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
			info+fullRE+": "+`\(hello|foo=world|bar, hello|foo=world|bar\) `+"testing\n",
		), nil)
	})

	t.Run("errorData color", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		logger := logging.Create(logging.OptWriter(out), logging.OptExit(func(int) {}), logging.OptColor())
		testRun(assert, "errorData color", logger, []any{"ErrorData", mockMap, "testing"}, out, regexp.MustCompile(
			ctrlSeqRE+erro+ctrlSeqRE+fullRE+": "+ctrlSeqRE+"hello"+ctrlSeqRE+"="+ctrlSeqRE+"world"+
				ctrlSeqRE+", "+ctrlSeqRE+"foo"+ctrlSeqRE+"="+ctrlSeqRE+"bar"+ctrlSeqRE+" testing",
		), nil)
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
		/* 1 Debug*/ {outREDebug, outREInfo, outREWarn, outREError, outREFatal},
		/* 2 Info*/ {nil, outREInfo, outREWarn, outREError, outREFatal},
		/* 3 Warn */ {nil, nil, outREWarn, outREError, outREFatal},
		/* 4 Error*/ {nil, nil, nil, outREError, outREFatal},
		/* 5 Fatal*/ {nil, nil, nil, nil, outREFatal},
		/* 6 None*/ {nil, nil, nil, nil, nil},
	}
	notExpected := [][]*regexp.Regexp{
		/* 1 Debug*/ {nil, nil, nil, nil, nil},
		/* 2 Info*/ {outREDebug, nil, nil, nil, nil},
		/* 3 Warn */ {outREDebug, outREInfo, nil, nil, nil},
		/* 4 Error*/ {outREDebug, outREInfo, outREWarn, nil, nil},
		/* 5 Fatal*/ {outREDebug, outREInfo, outREWarn, outREError, nil},
		/* 6 None*/ {outREDebug, outREInfo, outREWarn, outREError, outREFatal},
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
				testRun(require.New(t), name, logger, call, out, expected[i][callN], notExpected[i][callN])
			}
		})
	}
}
