package misc

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// blocks until receiving interrupt, hang up or terminate; then calls
// the handler function once.
func RoutineListenStopSignal(ctx context.Context, handler func(sig os.Signal)) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case s := <-signalChan:
		handler(s)
	case <-ctx.Done():
	}
}
