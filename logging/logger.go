// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package logging

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/tcodes0/go/hue"
)

const (
	appendEquals = "="
	appendSuffix = ", "
)

// Logger has no public fields; wraps log.Logger with additional functionality.
type Logger struct {
	l         *log.Logger
	mu        *sync.Mutex    // use Logger wrapper methods to avoid panic
	exit      func(code int) // proxy to os.Exit(1)
	metadata  string         // added to all messages
	level     Level          // only log if message level is >= to this
	msgLevel  Level          // level of the next message
	color     bool           // print terminal color characters
	calldepth int            // stack depth for log.Lshortfile
}

// used internally to control level prefix.
func (logger *Logger) setPrefix(prefix string) {
	if logger.l == nil {
		return
	}

	if logger.color {
		switch prefix {
		case info:
			prefix = hue.Cprint(hue.Gray, info)
		case warn:
			// warn and below: hue.Gray is added to color the log line information
			prefix = hue.Cprint(hue.Yellow, warn, hue.TermEnd, hue.Cprint(hue.Gray))
		case erro:
			prefix = hue.Cprint(hue.Red, erro, hue.TermEnd, hue.Cprint(hue.Gray))
		case fatal:
			prefix = hue.Cprint(hue.BrightRed, fatal, hue.TermEnd, hue.Cprint(hue.Gray))
		case debug:
			prefix = hue.Cprint(hue.Blue, debug, hue.TermEnd, hue.Cprint(hue.Gray))
		}
	}

	logger.l.SetPrefix(prefix)
}

// set a logger in this context, retrieve it with FromContext.
func (logger *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

// set Level of the next message to warning.
func (logger *Logger) Warn() *Logger {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	logger.setPrefix(warn)
	logger.msgLevel = LWarn

	return logger
}

// set Level of the next message to error.
func (logger *Logger) Error() *Logger {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	logger.setPrefix(erro)
	logger.msgLevel = LError

	return logger
}

// set Level of the next message to debug.
func (logger *Logger) Debug() *Logger {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	logger.setPrefix(debug)
	logger.msgLevel = LDebug

	return logger
}

// send a message.
func (logger *Logger) Log(msg ...any) {
	defer func() {
		logger.setPrefix(info)
		logger.calldepth = defaultCalldepth
		logger.msgLevel = LInfo
	}()

	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	if logger.l == nil || logger.msgLevel < logger.level {
		return
	}

	if logger.metadata != "" {
		msgMetadata := make([]any, len(msg)+1)

		if logger.color {
			msgMetadata[0] = strings.TrimSuffix(logger.metadata, appendSuffix) + hue.TermEnd + " "
		} else {
			msgMetadata[0] = "(" + strings.TrimSuffix(logger.metadata, appendSuffix) + ")" + " "
		}

		copy(msgMetadata[1:], msg)
		msg = msgMetadata
	}

	out := fmt.Sprint(msg...)
	if logger.color {
		// end color of the log line information, started on prefix
		out = hue.TermEnd + out
	}

	err := logger.l.Output(logger.calldepth, out)
	if err != nil {
		logger.l.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}
}

// send a formatted message.
func (logger *Logger) Logf(format string, args ...any) {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	out := fmt.Sprintf(format, args...)

	logger.calldepth++

	logger.Log(out)
}

// sends a message and then calls Logger.exit().
func (logger *Logger) Fatal(msg ...any) {
	// doesn't matter since we exit(1)
	_ = logger.lock()

	logger.setPrefix(fatal)

	logger.calldepth++
	logger.msgLevel = LFatal

	logger.Log(msg...)

	if logger.exit != nil {
		logger.exit(1)
	}
}

// sends a formatted message and then calls Logger.exit().
func (logger *Logger) Fatalf(format string, msg ...any) {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	out := fmt.Sprintf(format, msg...)

	logger.calldepth++

	logger.Fatal(out)
}

// append metadata to all future messages,
// metadata is formated in key value pairs;
// see Wipe.
func (logger *Logger) Metadata(key string, val any) *Logger {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	formatVal := fmt.Sprintf("%v", val)

	if logger.color {
		logger.metadata += hue.Cprint(hue.Brown, key) + hue.Cprint(hue.Gray, appendEquals) +
			hue.Cprint(hue.Brown, formatVal) + hue.Cprint(hue.Gray, appendSuffix)
	} else {
		logger.metadata += key + appendEquals + formatVal + appendSuffix
	}

	return logger
}

// remove all metadata from future messages,
// see Metadata.
func (logger *Logger) Wipe() *Logger {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	logger.metadata = ""

	return logger
}

// sets the function to call from Logger.Fatal methods.
func (logger *Logger) SetExit(exitFunc func(int)) {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	logger.exit = exitFunc
}

// set the level of the logger, messages < Logger.level will be ignored.
func (logger *Logger) SetLevel(level Level) {
	ok := logger.lock()
	if ok {
		defer logger.unlock()
	}

	logger.level = level
}

// lock the logger mutex with TryLock.
func (logger *Logger) lock() bool {
	if logger.mu == nil {
		return false
	}

	return logger.mu.TryLock()
}

// unlock the logger mutex.
func (logger *Logger) unlock() {
	if logger.mu == nil {
		return
	}

	logger.mu.Unlock()
}
