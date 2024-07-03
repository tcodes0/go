// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package logging

import (
	"context"
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/tcodes0/go/hue"
)

// Setting a level prevents messages with a lower level from being logged.
// The default level is LInfo, suppressing debug messages.
type Level int

const (
	// lowest level.
	LDebug Level = iota + 1
	// default level.
	LInfo
	LWarn
	LError
	LFatal
	// highest level.
	LNone

	debug string = "DEBUG "
	info  string = "INFO  "
	warn  string = "WARN  "
	erro  string = "ERROR "
	fatal string = "FATAL "

	// UTC time with file and line of log message.
	defaultFlags = log.LstdFlags | log.Lshortfile | log.LUTC

	equals = "="
	comma = ", "
)

type ContextKey struct{}

var (
	contextKey = ContextKey{}
	infoColor  = hue.Printc(hue.Gray, info)
	// warn and higher: hue.Gray is added to color the log line information.
	warnColor  = hue.Printc(hue.Yellow, warn, hue.End) + hue.Printc(hue.Gray)
	erroColor  = hue.Printc(hue.Red, erro, hue.End) + hue.Printc(hue.Gray)
	fatalColor = hue.Printc(hue.BrightRed, fatal, hue.End) + hue.Printc(hue.Gray)
	debugColor = hue.Printc(hue.Blue, debug, hue.End) + hue.Printc(hue.Gray)
)

// retrieves a logger from a context, see Logger.WithContext.
func FromContext(ctx context.Context) *Logger {
	l, ok := ctx.Value(contextKey).(*Logger)
	if !ok {
		panic("no logger found in context")
	}

	return l
}

type createOpts = struct {
	writer io.Writer
	exit   func(code int)
	level  Level
	flags  int
	color  bool
}

// functional options for creating a logger.
type CreateOptions func(c *createOpts)

// option to set flags for the logger.
func OptFlags(flags int) CreateOptions {
	return func(c *createOpts) {
		c.flags = flags
	}
}

// option to enable color output.
func OptColor() CreateOptions {
	return func(c *createOpts) {
		c.color = true
	}
}

// option to set the writer for the logger.
func OptWriter(w io.Writer) CreateOptions {
	return func(c *createOpts) {
		c.writer = w
	}
}

// option to set the exit function for the logger.
// useful for testing.
func OptExit(exit func(code int)) CreateOptions {
	return func(c *createOpts) {
		c.exit = exit
	}
}

// option to set the log level for the logger.
func OptLevel(level Level) CreateOptions {
	return func(c *createOpts) {
		c.level = level
	}
}

// creates a new logger with the given options.
func Create(options ...CreateOptions) *Logger {
	opts := &createOpts{
		flags:  defaultFlags,
		color:  false,
		writer: log.Writer(),
		exit:   os.Exit,
		level:  LInfo,
	}

	for _, o := range options {
		o(opts)
	}

	prefix := info
	if opts.color {
		prefix = hue.Printc(hue.Gray, info)
	}

	logger := &Logger{
		l:        log.New(opts.writer, prefix, opts.flags),
		color:    atomic.Bool{},
		exitFunc: opts.exit,
	}
	logger.color.Store(opts.color)

	return logger
}
