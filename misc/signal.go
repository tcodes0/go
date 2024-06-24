// Copyright 2024 Raphael Thomazella. All rights reserved.
//  Use of this source code is governed by the BSD-3-Clause
//  license that can be found in the LICENSE file and online
//  at https://opensource.org/license/BSD-3-clause.

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
