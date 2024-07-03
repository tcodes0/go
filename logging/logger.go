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
	"sync/atomic"

	"github.com/tcodes0/go/hue"
)

// Logger has no public fields; wraps log.Logger with additional functionality.
type Logger struct {
	l        *log.Logger
	exitFunc func(code int) // proxy to os.Exit(1)
	atmLevel atomic.Int32   // only log if message level is >= to this
	color    atomic.Bool    // print terminal color characters
}

// set a logger in this context, retrieve it with FromContext.
func (logger *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

func (logger *Logger) Info(msg ...any) {
	logger.out(LInfo, msg...)
}

func (logger *Logger) Infof(format string, args ...any) {
	logger.out(LInfo, fmt.Sprintf(format, args...))
}

func (logger *Logger) Warn(msg ...any) {
	logger.out(LWarn, msg...)
}

func (logger *Logger) Warnf(format string, args ...any) {
	logger.out(LWarn, fmt.Sprintf(format, args...))
}

func (logger *Logger) Error(msg ...any) {
	logger.out(LError, msg...)
}

func (logger *Logger) Errorf(format string, args ...any) {
	logger.out(LError, fmt.Sprintf(format, args...))
}

func (logger *Logger) Debug(msg ...any) {
	logger.out(LDebug, msg...)
}

func (logger *Logger) Debugf(format string, args ...any) {
	logger.out(LDebug, fmt.Sprintf(format, args...))
}

func (logger *Logger) Fatal(msg ...any) {
	logger.out(LFatal, msg...)
	logger.exit()
}

func (logger *Logger) Fatalf(format string, args ...any) {
	logger.out(LFatal, fmt.Sprintf(format, args...))
	logger.exit()
}

func (logger *Logger) exit() {
	if logger.exitFunc != nil {
		logger.exitFunc(1)
	}
}

func (logger *Logger) out(msgLevel Level, msg ...any) {
	if logger.l == nil || msgLevel < Level(logger.atmLevel.Load()) {
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
	// 2 for logger.l implementation + 1 for out() itself
	// + 1 for caller function
	calldepth := 4

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

	// Create a slice to hold stack trace info
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, allGoroutines)
	stackBuffer.Write(stack[:n])

	// Print to standard error (default logger output point)
	logger.out(LError, stackBuffer.String())
}

// set the level of the logger, messages < Logger.level will be ignored.
func (logger *Logger) SetLevel(level Level) {
	logger.atmLevel.Store(int32(level))
}

// append metadata to all future messages,
// metadata is formated in key value pairs;
// see Wipe.
// TODO: implement metadata.
func (logger *Logger) Metadata(key string, val any) *Logger {
	return logger
}
