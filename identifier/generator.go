// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package identifier

import (
	"context"
	"errors"
)

type ContextKey struct{}

var contextKey = ContextKey{}

type Generator interface {
	Generate() string
	WithContext(ctx context.Context) context.Context
}

// retrieves a generator from the context.
func FromContext(ctx context.Context) (Generator, error) {
	rand, ok := ctx.Value(contextKey).(Generator)
	if !ok {
		return nil, errors.New("no value found in context")
	}

	return rand, nil
}
