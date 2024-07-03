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

//nolint:funlen,gocognit,maintidx // test
func TestLogger(t *testing.T) {
	t.Parallel()

	regExpDate := `\d{4}/\d{2}/\d{2}`
	regExpTime := `\d{2}:\d{2}:\d{2}`
	regExpFileLine := `[a-z_]+\.go:\d+`
	// sequences of control chars to end or start formatting
	// matches one sequence after the other
	ctrlSeqRegExp := `[0-9;\033\[m]+`
	fullRegExp := regExpDate + " " + regExpTime + " " + regExpFileLine
	debug := "DEBUG "
	info := "INFO "
	warn := "WARN "
	erro := "ERROR "
	fatal := "FATAL "
	outREDebug := regexp.MustCompile(debug + fullRegExp + ": testing\n")
	outREInfo := regexp.MustCompile(info + fullRegExp + ": testing\n")
	outREWarn := regexp.MustCompile(warn + fullRegExp + ": testing\n")
	outREError := regexp.MustCompile(erro + fullRegExp + ": testing\n")
	outREFatal := regexp.MustCompile(fatal + fullRegExp + ": testing\n")
	levelCalls := [][]string{
		{"Debug", "testing"},
		{"Info", "testing"},
		{"Warn", "testing"},
		{"Error", "testing"},
		{"Fatal", "testing"},
	}
	levelRetTypes := [][]string{{}, {}, {}, {}, {}}
	map1 := map[string]any{
		"hello": "world",
		"foo":   "bar",
	}

	tests := []struct {
		name        string
		calls       [][]string
		retType     [][]string
		outMatch    []*regexp.Regexp
		outNotMatch []*regexp.Regexp
		opts        []logging.CreateOptions
		nop         bool
	}{
		{
			name:     "info log",
			calls:    [][]string{{"Info", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(info + fullRegExp + ": testing\n")},
		},
		{
			name:     "warn log",
			calls:    [][]string{{"Warn", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(warn + fullRegExp + ": testing\n")},
		},
		{
			name:     "warn log on nop logger",
			nop:      true,
			calls:    [][]string{{"Warn", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("^$")},
		},
		{
			name:     "error log formatted log",
			calls:    [][]string{{"Errorf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(erro + fullRegExp + ": testing\n")},
		},
		{
			name:     "debug log formatted log",
			calls:    [][]string{{"Debugf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(debug + fullRegExp + ": testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:     "fatal formatted log",
			calls:    [][]string{{"Fatalf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(fatal + fullRegExp + ": testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:     "fatal formatted log on nop logger",
			nop:      true,
			calls:    [][]string{{"Fatalf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("^$")},
		},
		{
			name:     "color info log",
			calls:    [][]string{{"Info", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqRegExp + info + fullRegExp + ": " + ctrlSeqRegExp + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "color warn log",
			calls:    [][]string{{"Warn", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqRegExp + warn + ctrlSeqRegExp + fullRegExp + ": " + ctrlSeqRegExp + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "color error formatted log",
			calls:    [][]string{{"Errorf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqRegExp + erro + ctrlSeqRegExp + fullRegExp + ": " + ctrlSeqRegExp + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "color debug formatted log",
			calls:    [][]string{{"Debugf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqRegExp + debug + ctrlSeqRegExp + fullRegExp + ": " + ctrlSeqRegExp + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug), logging.OptColor()},
		},
		{
			name:     "color fatal formatted log",
			calls:    [][]string{{"Fatalf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqRegExp + fatal + ctrlSeqRegExp + fullRegExp + ": " + ctrlSeqRegExp + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug), logging.OptColor()},
		},
		{
			name:     "data",
			calls:    [][]string{{"ErrorData", "<map1>", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(erro + fullRegExp + ": " + `\(hello|foo=world|bar, hello|foo=world|bar\) ` + "testing\n")},
		},
		{
			name:    "data color",
			calls:   [][]string{{"ErrorData", "<map1>", "testing"}},
			retType: [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(
				ctrlSeqRegExp + erro + ctrlSeqRegExp + fullRegExp + ": " + ctrlSeqRegExp + "hello" + ctrlSeqRegExp + "=" + ctrlSeqRegExp + "world" +
					ctrlSeqRegExp + ", " + ctrlSeqRegExp + "foo" + ctrlSeqRegExp + "=" + ctrlSeqRegExp + "bar" + ctrlSeqRegExp + " testing",
			)},
			opts: []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "debug level",
			calls:    levelCalls,
			retType:  levelRetTypes,
			outMatch: []*regexp.Regexp{outREDebug, outREInfo, outREWarn, outREError, outREFatal},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:        "info level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outREInfo, outREWarn, outREError, outREFatal},
			outNotMatch: []*regexp.Regexp{outREDebug},
		},
		{
			name:        "warn level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outREWarn, outREError, outREFatal},
			outNotMatch: []*regexp.Regexp{outREDebug, outREInfo},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LWarn)},
		},
		{
			name:        "error level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outREError, outREFatal},
			outNotMatch: []*regexp.Regexp{outREDebug, outREInfo, outREWarn},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LError)},
		},
		{
			name:        "fatal level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outREFatal},
			outNotMatch: []*regexp.Regexp{outREDebug, outREInfo, outREWarn, outREError},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LFatal)},
		},
		{
			name:        "none level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outNotMatch: []*regexp.Regexp{outREDebug, outREInfo, outREWarn, outREError, outREFatal},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LNone)},
		},
	}

	assert := require.New(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.Buffer{}
			test.opts = append(test.opts, logging.OptWriter(&buf), logging.OptExit(func(int) {}))
			logger := &logging.Logger{}

			if !test.nop {
				logger = logging.Create(test.opts...)
			}

			for callN, call := range test.calls {
				method := reflect.ValueOf(logger).MethodByName(call[0])

				if !method.IsValid() {
					assert.Fail("invalid method", "call [%d] method %s not found", callN, call[0])
				}

				args := make([]reflect.Value, len(call)-1)

				for i, arg := range call[1:] {
					if arg == "<map1>" {
						args[i] = reflect.ValueOf(map1)
					} else {
						args[i] = reflect.ValueOf(arg)
					}
				}

				returns := method.Call(args)
				assert.Len(returns, len(test.retType[callN]), fmt.Sprintf("unexpected return values on call [%d]", callN))

				for i, ret := range returns {
					assert.Equal(ret.Type().String(), test.retType[callN][i], fmt.Sprintf("unexpected return type at [%d] on call [%d]", i, callN))
				}
			}

			for _, reg := range test.outMatch {
				assert.Regexp(reg, buf.String(), "unexpected output on test '%s'", test.name)
			}

			for _, reg := range test.outNotMatch {
				assert.NotRegexp(reg, buf.String(), "expected output on test '%s'", test.name)
			}
		})
	}
}

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
