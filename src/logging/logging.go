package logging

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type Level string

const (
	Debug Level = "debug"
	Warn  Level = "warn"
	Error Level = "error"
)

func (l Level) String() string {
	return string(l)
}

// should call only once. Sets a context logger
func Init(ctx context.Context) (context.Context, zerolog.Logger) {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	return logger.WithContext(ctx), logger
}

func SetGlobalLevel(l string) {
	switch Level(l) {
	case Debug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case Warn:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case Error:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// returns a non-nil logger, it may be disabled if not found on CTX
func FromContext(ctx context.Context) zerolog.Logger {
	if ctxLogger := zerolog.Ctx(ctx); ctxLogger != nil {
		return *ctxLogger
	}

	return zerolog.Nop()
}
