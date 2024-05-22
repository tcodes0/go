package logging

import (
	"context"
	"errors"
	"log"
	"os"
)

type Level int

const (
	LDebug Level = iota + 1
	LInfo  Level = iota + 1
	LWarn  Level = iota + 1
	LError Level = iota + 1

	info  string = "INFO: " // default level
	warn  string = "WARN: "
	erro  string = "ERRO: "
	fatal string = "FATL: "
	debug string = "DEBG: "

	defaultFlags     = log.LstdFlags | log.Lshortfile | log.LUTC
	defaultCalldepth = 2 // necessary for log.Lshortfile to show correctly
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
		prefix = gray(info)
	}

	return &Logger{
		l:         log.New(log.Writer(), prefix, flags),
		level:     level,
		color:     color,
		calldepth: defaultCalldepth,
		Exit: func(code int) {
			os.Exit(code)
		},
	}
}

func Nop() *Logger {
	return &Logger{}
}
