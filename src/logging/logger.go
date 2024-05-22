package logging

import (
	"context"
	"fmt"
	"log"
	"os"
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
	logger.SetPrefix(warn)

	return logger
}

func (logger *Logger) Error() *Logger {
	logger.SetPrefix(erro)

	return logger
}

func (logger *Logger) Debug() *Logger {
	logger.SetPrefix(debug)

	return logger
}

func (logger *Logger) Log(v ...interface{}) {
	s := make([]interface{}, len(v)+1)
	s[0] = logger.appended + ": "
	copy(s[1:], v)

	out := fmt.Sprint(v...)

	err := logger.l.Output(calldepth, out)
	if err != nil {
		logger.l.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}

	logger.SetPrefix(info)
}

func (logger *Logger) Logf(format string, v ...interface{}) {
	out := fmt.Sprintf(logger.appended+": "+format, v...)

	err := logger.l.Output(calldepth, out)
	if err != nil {
		logger.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}

	logger.SetPrefix(info)
}

func (logger *Logger) Fatal(msg ...interface{}) {
	s := make([]interface{}, len(msg)+1)
	s[0] = logger.appended + ": "
	copy(s[1:], msg)

	logger.SetPrefix(fatal)

	out := fmt.Sprint(msg...)

	err := logger.l.Output(calldepth, out)
	if err != nil {
		logger.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}

	os.Exit(1)
}

func (logger *Logger) Fatalf(format string, msg ...interface{}) {
	logger.SetPrefix(fatal)

	out := fmt.Sprintf(logger.appended+": "+format, msg...)

	err := logger.l.Output(calldepth, out)
	if err != nil {
		logger.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}

	os.Exit(1)
}

func (logger *Logger) Append(message string) {
	logger.appended += message + " "
}

func (logger *Logger) Unappend() {
	logger.appended = ""
}

func (logger *Logger) SetPrefix(prefix string) {
	if logger.color {
		switch prefix {
		case info:
			prefix = gray(info)
		case warn:
			prefix = yellow(warn)
		case erro:
			prefix = red(erro)
		case fatal:
			prefix = darkRed(fatal)
		case debug:
			prefix = blue(debug)
		}
	}

	logger.l.SetPrefix(prefix)
}
