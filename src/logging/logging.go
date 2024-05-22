package logging

import (
	"context"
	"errors"
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

	info  string = "INFO " // default level
	warn  string = "WARN "
	erro  string = "ERRO "
	fatal string = "FATL "
	debug string = "DEBG "

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

func FromContext(ctx context.Context) (*Logger, error) {
	l, ok := ctx.Value(contextKey).(*Logger)
	if !ok {
		return nil, ErrNoCtxLogger
	}

	return l, nil
}

func Create(level Level, flags int, color bool) *Logger {
	if flags == 0 {
		flags = defaultFlags
	}

	prefix := info
	if color {
		prefix = hue.Cprint(hue.Gray, info)
	}

	return &Logger{
		l:         log.New(log.Writer(), prefix, flags),
		level:     level,
		color:     color,
		calldepth: defaultCalldepth,
		exit: func(code int) {
			os.Exit(code)
		},
	}
}
