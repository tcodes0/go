// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package logging

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/tcodes0/go/hue"
)

// logger wraps log.Logger.
type Logger struct {
	l        *log.Logger
	exitFunc func(code int) // proxy to os.Exit(1)
	level    atomic.Int32   // messages are ignored if their level is less
	color    atomic.Bool    // print terminal color characters
}

// set a logger in this context, retrieve it with FromContext.
func (logger *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

// sends a message with level info.
func (logger *Logger) Info(msg ...any) {
	logger.out(LInfo, msg...)
}

// sends a formatted message with level info.
func (logger *Logger) Infof(format string, args ...any) {
	logger.out(LInfo, fmt.Sprintf(format, args...))
}

// sends a message with level info and data. Note data ordering is not stable.
func (logger *Logger) InfoData(data map[string]any, msg ...any) {
	logger.out(LInfo, slices.Concat([]any{logger.format(data)}, msg)...)
}

// sends a message with level warn.
func (logger *Logger) Warn(msg ...any) {
	logger.out(LWarn, msg...)
}

// sends a formatted message with level warn.
func (logger *Logger) Warnf(format string, args ...any) {
	logger.out(LWarn, fmt.Sprintf(format, args...))
}

// sends a message with level warn and data. Note data ordering is not stable.
func (logger *Logger) WarnData(data map[string]any, msg ...any) {
	logger.out(LWarn, slices.Concat([]any{logger.format(data)}, msg)...)
}

// sends a message with level error.
func (logger *Logger) Error(msg ...any) {
	logger.out(LError, msg...)
}

// sends a formatted message with level error.
func (logger *Logger) Errorf(format string, args ...any) {
	logger.out(LError, fmt.Sprintf(format, args...))
}

// sends a message with level error and data. Note data ordering is not stable.
func (logger *Logger) ErrorData(data map[string]any, msg ...any) {
	logger.out(LError, slices.Concat([]any{logger.format(data)}, msg)...)
}

// sends a message with level debug.
func (logger *Logger) Debug(msg ...any) {
	logger.out(LDebug, msg...)
}

// sends a formatted message with level debug.
func (logger *Logger) Debugf(format string, args ...any) {
	logger.out(LDebug, fmt.Sprintf(format, args...))
}

// sends a message with level debug and data. Note data ordering is not stable.
func (logger *Logger) DebugData(data map[string]any, msg ...any) {
	logger.out(LDebug, slices.Concat([]any{logger.format(data)}, msg)...)
}

// sends a message with level fatal and calls the logger exit function.
func (logger *Logger) Fatal(msg ...any) {
	logger.out(LFatal, msg...)
	logger.exit()
}

// sends a formatted message with level fatal and calls the logger exit function.
func (logger *Logger) Fatalf(format string, args ...any) {
	logger.out(LFatal, fmt.Sprintf(format, args...))
	logger.exit()
}

// sends a message with level fatal, data, and calls the logger exit function.
func (logger *Logger) FatalData(data map[string]any, msg ...any) {
	logger.out(LFatal, slices.Concat([]any{logger.format(data)}, msg)...)
}

func (logger *Logger) format(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	color := logger.color.Load()
	dataMsg := ""

	for key, val := range data {
		sVal := fmt.Sprintf("%v", val)

		if color {
			dataMsg += hue.Printc(hue.Brown, key) + hue.Printc(hue.Gray, equals) +
				hue.Printc(hue.Brown, sVal) + hue.Printc(hue.Gray, comma)
		} else {
			dataMsg += key + equals + sVal + comma
		}
	}

	if color {
		return strings.TrimSuffix(dataMsg, comma) + hue.End + " "
	}

	return "(" + strings.TrimSuffix(dataMsg, comma) + ")" + " "
}

func (logger *Logger) exit() {
	if logger.exitFunc != nil {
		logger.exitFunc(1)
	}
}

func (logger *Logger) out(msgLevel Level, msg ...any) {
	if logger.l == nil || msgLevel < Level(logger.level.Load()) {
		return
	}

	out := fmt.Sprint(msg...)

	color := logger.color.Load()
	if color {
		// end color of the log line information, started on prefix
		out = hue.End + out
	}

	//nolint:exhaustive // default handles LInfo and return above on LNone
	switch msgLevel {
	default:
		if color {
			logger.setPrefix(infoColor)
		} else {
			logger.setPrefix(info)
		}
	case LWarn:
		if color {
			logger.setPrefix(warnColor)
		} else {
			logger.setPrefix(warn)
		}
	case LError:
		if color {
			logger.setPrefix(erroColor)
		} else {
			logger.setPrefix(erro)
		}
	case LFatal:
		if color {
			logger.setPrefix(fatalColor)
		} else {
			logger.setPrefix(fatal)
		}
	case LDebug:
		if color {
			logger.setPrefix(debugColor)
		} else {
			logger.setPrefix(debug)
		}
	}

	// controls stack frames to pop when showing file:line.
	// necessary for log.Lshortfile to show correctly
	// 2 lib defined + 1 for caller function
	calldepth := 3

	err := logger.l.Output(calldepth, out)
	if err != nil {
		logger.l.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}
}

func (logger *Logger) setPrefix(prefix string) {
	if logger.l.Prefix() != prefix {
		logger.l.SetPrefix(prefix)
	}
}

// prints a stacktrace as an error level log.
func (logger *Logger) Stacktrace(allGoroutines bool) {
	var stackBuffer bytes.Buffer

	stack := make([]byte, 2048)
	n := runtime.Stack(stack, allGoroutines)
	stackBuffer.Write(stack[:n])

	logger.out(LError, stackBuffer.String())
}

// set the level of the logger, lesser messages will be ignored.
func (logger *Logger) SetLevel(level Level) {
	logger.level.Store(int32(level))
}
