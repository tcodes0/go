// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package clock

import (
	"context"
	"time"
)

// Time implements nower with runtime time and timezone.
type Time struct {
	Location time.Location
}

var _ Nower = (*Time)(nil)

// returns the current time in timezone.
func (t *Time) Now() time.Time {
	return time.Now().In(&t.Location)
}

// returns a context with Time.
func (t *Time) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, t)
}
