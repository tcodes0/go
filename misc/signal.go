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
