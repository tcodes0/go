// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package logging_test

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/logging"
)

//nolint:funlen,maintidx // test
func TestLogger(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	regExpDate := `\d{4}/\d{2}/\d{2}`
	regExpTime := `\d{2}:\d{2}:\d{2}`
	regExpFileLine := `[a-z_]+\.go:\d+`
	// sequences of control chars to end or start formatting
	// matches one sequence after the other
	ctrlSeqs := `[0-9;\033\[m]+`
	fullRegExp := regExpDate + " " + regExpTime + " " + regExpFileLine
	retLogger := "*logging.Logger"
	debug := "DEBUG "
	info := "INFO  "
	warn := "WARN  "
	erro := "ERROR "
	fatal := "FATAL "
	outRegDebug := regexp.MustCompile(debug + fullRegExp + ": testing\n")
	outRegInfo := regexp.MustCompile(info + fullRegExp + ": testing\n")
	outRegWarn := regexp.MustCompile(warn + fullRegExp + ": testing\n")
	outRegError := regexp.MustCompile(erro + fullRegExp + ": testing\n")
	outRegFatal := regexp.MustCompile(fatal + fullRegExp + ": testing\n")

	levelCalls := [][]string{
		{"Log", "testing"},
		{"Fatal", "testing"},
		{"Debug"},
		{"Log", "testing"},
		{"Warn"},
		{"Log", "testing"},
		{"Error"},
		{"Log", "testing"},
	}
	levelRetTypes := [][]string{
		{},
		{},
		{retLogger},
		{},
		{retLogger},
		{},
		{retLogger},
		{},
		{retLogger},
		{},
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
			calls:    [][]string{{"Log", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(info + fullRegExp + ": testing\n")},
		},
		{
			name:     "warn log",
			calls:    [][]string{{"Warn"}, {"Log", "testing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(warn + fullRegExp + ": testing\n")},
		},
		{
			name:     "warn log on nop logger",
			nop:      true,
			calls:    [][]string{{"Warn"}, {"Log", "testing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("^$")},
		},
		{
			name:     "error logf",
			calls:    [][]string{{"Error"}, {"Logf", "test%s", "ing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(erro + fullRegExp + ": testing\n")},
		},
		{
			name:     "debug logf",
			calls:    [][]string{{"Debug"}, {"Logf", "test%s", "ing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(debug + fullRegExp + ": testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:     "fatalf",
			calls:    [][]string{{"Fatalf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(fatal + fullRegExp + ": testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:     "fatalf on nop logger",
			nop:      true,
			calls:    [][]string{{"Fatalf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("^$")},
		},
		{
			name:     "color info log",
			calls:    [][]string{{"Log", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqs + info + fullRegExp + ": " + ctrlSeqs + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "color warn log",
			calls:    [][]string{{"Warn"}, {"Log", "testing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqs + warn + ctrlSeqs + fullRegExp + ": " + ctrlSeqs + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "color error logf",
			calls:    [][]string{{"Error"}, {"Logf", "test%s", "ing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqs + erro + ctrlSeqs + fullRegExp + ": " + ctrlSeqs + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "color debug logf",
			calls:    [][]string{{"Debug"}, {"Logf", "test%s", "ing"}},
			retType:  [][]string{{retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqs + debug + ctrlSeqs + fullRegExp + ": " + ctrlSeqs + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug), logging.OptColor()},
		},
		{
			name:     "color fatalf",
			calls:    [][]string{{"Fatalf", "test%s", "ing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(ctrlSeqs + fatal + ctrlSeqs + fullRegExp + ": " + ctrlSeqs + "testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug), logging.OptColor()},
		},
		{
			name: "one metadata",
			calls: [][]string{
				{"Metadata", "hello", "world"},
				{"Error"},
				{"Logf", "test%s", "ing"},
			},
			retType:  [][]string{{retLogger}, {retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(erro + fullRegExp + ": " + `\(hello=world\) ` + "testing\n")},
		},
		{
			name: "metadata wipe",
			calls: [][]string{
				{"Metadata", "hello", "world"},
				{"Wipe"},
				{"Error"},
				{"Logf", "test%s", "ing"},
			},
			retType:  [][]string{{retLogger}, {retLogger}, {retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(erro + fullRegExp + ": " + "testing\n")},
		},
		{
			name: "many metadata",
			calls: [][]string{
				{"Metadata", "hello", "world"},
				{"Metadata", "foo", "bar"},
				{"Error"},
				{"Logf", "test%s", "ing"},
			},
			retType:  [][]string{{retLogger}, {retLogger}, {retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(erro + fullRegExp + ": " + `\(hello=world, foo=bar\) ` + "testing\n")},
		},
		{
			name: "many metadata color",
			calls: [][]string{
				{"Metadata", "hello", "world"},
				{"Metadata", "foo", "bar"},
				{"Error"},
				{"Logf", "test%s", "ing"},
			},
			retType: [][]string{{retLogger}, {retLogger}, {retLogger}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile(
				ctrlSeqs + erro + ctrlSeqs + fullRegExp + ": " + ctrlSeqs + "hello" + ctrlSeqs + "=" + ctrlSeqs + "world" + ctrlSeqs + ", " +
					ctrlSeqs + "foo" + ctrlSeqs + "=" + ctrlSeqs + "bar" + ctrlSeqs + " testing",
			)},
			opts: []logging.CreateOptions{logging.OptColor()},
		},
		{
			name:     "debug level",
			calls:    levelCalls,
			retType:  levelRetTypes,
			outMatch: []*regexp.Regexp{outRegDebug, outRegInfo, outRegWarn, outRegError, outRegFatal},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:        "info level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outRegInfo, outRegWarn, outRegError, outRegFatal},
			outNotMatch: []*regexp.Regexp{outRegDebug},
		},
		{
			name:        "warn level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outRegWarn, outRegError, outRegFatal},
			outNotMatch: []*regexp.Regexp{outRegDebug, outRegInfo},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LWarn)},
		},
		{
			name:        "error level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outRegError, outRegFatal},
			outNotMatch: []*regexp.Regexp{outRegDebug, outRegInfo, outRegWarn},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LError)},
		},
		{
			name:        "fatal level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outMatch:    []*regexp.Regexp{outRegFatal},
			outNotMatch: []*regexp.Regexp{outRegDebug, outRegInfo, outRegWarn, outRegError},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LFatal)},
		},
		{
			name:        "none level",
			calls:       levelCalls,
			retType:     levelRetTypes,
			outNotMatch: []*regexp.Regexp{outRegDebug, outRegInfo, outRegWarn, outRegError, outRegFatal},
			opts:        []logging.CreateOptions{logging.OptLevel(logging.LNone)},
		},
	}

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
					args[i] = reflect.ValueOf(arg)
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
