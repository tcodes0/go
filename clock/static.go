// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package clock

import (
	"context"
	time "time"
)

// Static implements nower with the same time returned every time.
type Static struct {
	Source time.Time
}

var _ Nower = (*Static)(nil)

// returns the same time every time.
func (s Static) Now() time.Time {
	return s.Source
}

// returns a context with Static.
func (s Static) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, s)
}
