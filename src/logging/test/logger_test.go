package test

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/logging"
)

//nolint:funlen // test
func TestLogger(t *testing.T) {
	assert := require.New(t)
	regExpDate := `\d{4}/\d{2}/\d{2}`
	regExpTime := `\d{2}:\d{2}:\d{2}`
	regExpFileLine := `[a-z_]+\.go:\d+`
	fullRegExp := regExpDate + " " + regExpTime + " " + regExpFileLine

	levelCalls := [][]string{
		{"Log", "testing"},
		{"Fatal", "testing"},
		//nolint:gofumpt // test
		{"Debug"}, {"Log", "testing"},
		{"Warn"}, {"Log", "testing"},
		{"Error"}, {"Log", "testing"},
	}
	levelRetTypes := [][]string{
		{},
		{},
		//nolint:gofumpt // test
		{"*logging.Logger"}, {},
		{"*logging.Logger"}, {},
		{"*logging.Logger"}, {},
		{"*logging.Logger"}, {},
	}

	tests := []struct {
		name        string
		calls       [][]string
		retType     [][]string
		outMatch    []*regexp.Regexp
		outNotMatch []*regexp.Regexp
		opts        []logging.CreateOptions
	}{
		{
			name:     "info log",
			calls:    [][]string{{"Log", "testing"}},
			retType:  [][]string{{}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("INFO " + fullRegExp + ": testing\n")},
		},
		{
			name:     "warn log",
			calls:    [][]string{{"Warn"}, {"Log", "testing"}},
			retType:  [][]string{{"*logging.Logger"}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("WARN " + fullRegExp + ": testing\n")},
		},
		{
			name:     "error log",
			calls:    [][]string{{"Error"}, {"Log", "testing"}},
			retType:  [][]string{{"*logging.Logger"}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("ERRO " + fullRegExp + ": testing\n")},
		},
		{
			name:     "debug log",
			calls:    [][]string{{"Debug"}, {"Log", "testing"}},
			retType:  [][]string{{"*logging.Logger"}, {}},
			outMatch: []*regexp.Regexp{regexp.MustCompile("DEBG " + fullRegExp + ": testing\n")},
			opts:     []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:    "debug level",
			calls:   levelCalls,
			retType: levelRetTypes,
			outMatch: []*regexp.Regexp{
				regexp.MustCompile("DEBG " + fullRegExp + ": testing\n"),
				regexp.MustCompile("INFO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("WARN " + fullRegExp + ": testing\n"),
				regexp.MustCompile("ERRO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("FATL " + fullRegExp + ": testing\n"),
			},
			opts: []logging.CreateOptions{logging.OptLevel(logging.LDebug)},
		},
		{
			name:    "info level",
			calls:   levelCalls,
			retType: levelRetTypes,
			outMatch: []*regexp.Regexp{
				regexp.MustCompile("INFO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("WARN " + fullRegExp + ": testing\n"),
				regexp.MustCompile("ERRO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("FATL " + fullRegExp + ": testing\n"),
			},
			outNotMatch: []*regexp.Regexp{
				regexp.MustCompile("DEBG " + fullRegExp + ": testing\n"),
			},
		},
		{
			name:    "warn level",
			calls:   levelCalls,
			retType: levelRetTypes,
			outMatch: []*regexp.Regexp{
				regexp.MustCompile("ERRO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("WARN " + fullRegExp + ": testing\n"),
				regexp.MustCompile("FATL " + fullRegExp + ": testing\n"),
			},
			outNotMatch: []*regexp.Regexp{
				regexp.MustCompile("INFO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("DEBG " + fullRegExp + ": testing\n"),
			},
			opts: []logging.CreateOptions{logging.OptLevel(logging.LWarn)},
		},
		{
			name:    "error level",
			calls:   levelCalls,
			retType: levelRetTypes,
			outMatch: []*regexp.Regexp{
				regexp.MustCompile("ERRO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("FATL " + fullRegExp + ": testing\n"),
			},
			outNotMatch: []*regexp.Regexp{
				regexp.MustCompile("WARN " + fullRegExp + ": testing\n"),
				regexp.MustCompile("INFO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("DEBG " + fullRegExp + ": testing\n"),
			},
			opts: []logging.CreateOptions{logging.OptLevel(logging.LError)},
		},
		{
			name:    "fatal level",
			calls:   levelCalls,
			retType: levelRetTypes,
			outMatch: []*regexp.Regexp{
				regexp.MustCompile("FATL " + fullRegExp + ": testing\n"),
			},
			outNotMatch: []*regexp.Regexp{
				regexp.MustCompile("ERRO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("WARN " + fullRegExp + ": testing\n"),
				regexp.MustCompile("INFO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("DEBG " + fullRegExp + ": testing\n"),
			},
			opts: []logging.CreateOptions{logging.OptLevel(logging.LFatal)},
		},
		{
			name:    "none level",
			calls:   levelCalls,
			retType: levelRetTypes,
			outNotMatch: []*regexp.Regexp{
				regexp.MustCompile("FATL " + fullRegExp + ": testing\n"),
				regexp.MustCompile("ERRO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("WARN " + fullRegExp + ": testing\n"),
				regexp.MustCompile("INFO " + fullRegExp + ": testing\n"),
				regexp.MustCompile("DEBG " + fullRegExp + ": testing\n"),
			},
			opts: []logging.CreateOptions{logging.OptLevel(logging.LNone)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// defer func() {
			// 	if r := recover(); r != nil {
			// 		fmt.Println("Recovered in f", r)
			// 	}
			// }()

			buf := bytes.Buffer{}
			test.opts = append(test.opts, logging.OptWriter(&buf), logging.OptExit(func(int) {}))

			logger := logging.Create(test.opts...)

			for callN, call := range test.calls {
				method := reflect.ValueOf(logger).MethodByName(call[0])

				if !method.IsValid() {
					assert.Fail("invalid method", "call [%d] method %s not found", callN, call[0])
				}

				args := make([]reflect.Value, len(call)-1)

				for i, arg := range call[1:] {
					args[i] = reflect.ValueOf(arg)
				}

				ret := method.Call(args)
				assert.Len(ret, len(test.retType[callN]), fmt.Sprintf("unexpected return values on call [%d]", callN))

				for i, r := range ret {
					assert.Equal(r.Type().String(), test.retType[callN][i], fmt.Sprintf("unexpected return type at [%d] on call [%d]", i, callN))
				}
			}

			for _, reg := range test.outMatch {
				assert.Regexp(reg, buf.String(), "unexpected output")
			}

			for _, reg := range test.outNotMatch {
				assert.NotRegexp(reg, buf.String(), "expected output")
			}
		})
	}
}
