package logging

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/tcodes0/go/src/hue"
)

type Level int

const (
	LDebug Level = iota + 1
	LInfo  Level = iota + 1
	LWarn  Level = iota + 1
	LError Level = iota + 1
	LFatal Level = iota + 1
	LNone  Level = iota + 1

	debug string = "DEBG "
	info  string = "INFO " // default level
	warn  string = "WARN "
	erro  string = "ERRO "
	fatal string = "FATL "

	defaultFlags = log.LstdFlags | log.Lshortfile | log.LUTC
	// necessary for log.Lshortfile to show correctly
	// controls stack frames to pop when showing file:line.
	defaultCalldepth = 2
)

type ContextKey struct{}

var (
	ErrNoCtxLogger = errors.New("no logger found in context")

	contextKey = ContextKey{}
)

// retrieves a logger from a context, see Logger.WithContext.
func FromContext(ctx context.Context) (*Logger, error) {
	l, ok := ctx.Value(contextKey).(*Logger)
	if !ok {
		return nil, ErrNoCtxLogger
	}

	return l, nil
}

type createOpts = struct {
	writer io.Writer
	exit   func(code int)
	level  Level
	flags  int
	color  bool
}

type CreateOptions func(c *createOpts)

func OptFlags(flags int) CreateOptions {
	return func(c *createOpts) {
		c.flags = flags
	}
}

func OptColor(color bool) CreateOptions {
	return func(c *createOpts) {
		c.color = color
	}
}

func OptWriter(w io.Writer) CreateOptions {
	return func(c *createOpts) {
		c.writer = w
	}
}

func OptExit(exit func(code int)) CreateOptions {
	return func(c *createOpts) {
		c.exit = exit
	}
}

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
