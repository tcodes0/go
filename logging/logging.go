package logging

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/tcodes0/go/hue"
)

// Setting a level prevents messages with a lower level from being logged.
// The default level is LInfo, suppressing debug messages.
type Level int

const (
	// lowest level.
	LDebug Level = iota + 1
	// default level.
	LInfo  Level = iota + 1
	LWarn  Level = iota + 1
	LError Level = iota + 1
	LFatal Level = iota + 1
	// highest level.
	LNone Level = iota + 1

	debug string = "DEBG "
	info  string = "INFO "
	warn  string = "WARN "
	erro  string = "ERRO "
	fatal string = "FATL "

	// UTC time with file and line of log message.
	defaultFlags = log.LstdFlags | log.Lshortfile | log.LUTC
	// necessary for log.Lshortfile to show correctly
	// controls stack frames to pop when showing file:line.
	defaultCalldepth = 2
)

type ContextKey struct{}

var contextKey = ContextKey{}

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
func OptColor(color bool) CreateOptions {
	return func(c *createOpts) {
		c.color = color
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
		prefix = hue.Cprint(hue.Gray, info)
	}

	return &Logger{
		l:         log.New(opts.writer, prefix, opts.flags),
		level:     opts.level,
		color:     opts.color,
		msgLevel:  LInfo,
		calldepth: defaultCalldepth,
		exit:      opts.exit,
	}
}
