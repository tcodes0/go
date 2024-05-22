package logging

import (
	"context"
	"log"
)

type Logger struct {
	l        *log.Logger
	appended string
	level    Level
	color    bool
}

func (logger *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

func (logger *Logger) Warn() *Logger {
	prefix := warn
	if logger.color {
		prefix = yellow(warn)
	}

	logger.l.SetPrefix(prefix)

	return logger
}

func (logger *Logger) Error() *Logger {
	prefix := erro
	if logger.color {
		prefix = red(erro)
	}

	logger.l.SetPrefix(prefix)

	return logger
}

func (logger *Logger) Debug() *Logger {
	prefix := debug
	if logger.color {
		prefix = blue(debug)
	}

	logger.l.SetPrefix(prefix)

	return logger
}

func (logger *Logger) Log(v ...interface{}) {
	s := make([]interface{}, len(v)+1)
	s[0] = logger.appended + ": "
	copy(s[1:], v)

	logger.l.Print(v...)

	prefix := info
	if logger.color {
		prefix = green(info)
	}

	logger.l.SetPrefix(prefix)
}

func (logger *Logger) Logf(format string, v ...interface{}) {
	logger.l.Printf(logger.appended+": "+format, v...)

	prefix := info
	if logger.color {
		prefix = green(info)
	}

	logger.l.SetPrefix(prefix)
}

func (logger *Logger) Fatal(v ...interface{}) {
	s := make([]interface{}, len(v)+1)
	s[0] = logger.appended + ": "
	copy(s[1:], v)

	logger.l.Fatal(v...)
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.l.Fatalf(logger.appended+": "+format, v...)
}

func (logger *Logger) Append(message string) {
	logger.appended += message + " "
}

func (logger *Logger) Unappend() {
	logger.appended = ""
}
