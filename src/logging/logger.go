package logging

import (
	"context"
	"fmt"
	"log"
	"strings"
)

const (
	appendEquals = "="
	appendSuffix = ", "
)

type Logger struct {
	l         *log.Logger
	Exit      func(code int) // proxy to os.Exit(1)
	appended  string         // add to all messages
	level     Level
	color     bool // print terminal color
	calldepth int  // track stack depth for log.Lshortfile
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

func (logger *Logger) Log(msg ...interface{}) {
	if logger.l == nil {
		return
	}

	if logger.appended != "" {
		appendedMsg := make([]interface{}, len(msg)+1)

		if logger.color {
			appendedMsg[0] = strings.TrimSuffix(logger.appended, appendSuffix) + ": " + colorEnd
		} else {
			appendedMsg[0] = strings.TrimSuffix(logger.appended, appendSuffix) + ": "
		}

		for i, m := range msg {
			appendedMsg[i+1] = m
		}

		msg = appendedMsg
	}

	out := fmt.Sprint(msg...)
	if logger.color {
		// end color of the log line information, started on prefix
		out = colorEnd + out
	}

	err := logger.l.Output(logger.calldepth, out)
	if err != nil {
		logger.l.SetPrefix(erro)
		logger.l.Print("printing previous log line: " + err.Error())
	}

	logger.SetPrefix(info)
	logger.calldepth = defaultCalldepth
}

func (logger *Logger) Logf(format string, v ...interface{}) {
	out := fmt.Sprintf(format, v...)

	logger.calldepth++

	logger.Log(out)
}

func (logger *Logger) Fatal(msg ...interface{}) {
	logger.SetPrefix(fatal)

	logger.calldepth++

	logger.Log(msg...)

	if logger.Exit != nil {
		logger.Exit(1)
	}
}

func (logger *Logger) Fatalf(format string, msg ...interface{}) {
	out := fmt.Sprintf(format, msg...)

	logger.calldepth++

	logger.Fatal(out)
}

func (logger *Logger) Append(key, val string) {
	if logger.color {
		logger.appended += lightGray(key) + gray(appendEquals) + lightGray(val) + gray(appendSuffix)
	} else {
		logger.appended += key + appendEquals + val + appendSuffix
	}
}

func (logger *Logger) Unappend() {
	logger.appended = ""
}

func (logger *Logger) SetPrefix(prefix string) {
	if logger.l == nil {
		return
	}

	if logger.color {
		switch prefix {
		case info:
			prefix = gray(info)
		case warn:
			prefix = yellow(warn)
		case erro:
			prefix = red(erro)
		case fatal:
			prefix = brightRed(fatal)
		case debug:
			prefix = blue(debug)
		}
	}

	logger.l.SetPrefix(prefix)
}
