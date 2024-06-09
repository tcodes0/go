// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package clock

import (
	"context"
	"errors"
	time "time"
)

type ContextKey struct{}

var contextKey = ContextKey{}

// an interface with Now that provides the current time.
type Nower interface {
	Now() time.Time
	WithContext(ctx context.Context) context.Context
}

// retrieves a Nower from a context.
func FromContext(ctx context.Context) (Nower, error) {
	nower, ok := ctx.Value(contextKey).(Nower)
	if !ok {
		return nil, errors.New("no value found in context")
	}

	return nower, nil
}
