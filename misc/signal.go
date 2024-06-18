package misc

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// catches system signals like interrupt, hang up and terminate once and calls
// the handler function.
func ListenStopSignal(ctx context.Context, handler func(sig os.Signal)) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case s := <-signalChan:
		handler(s)
	case <-ctx.Done():
	}
}
