// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package identifier

import (
	"context"

	"github.com/google/uuid"
)

// This concrete type implements the Generator interface.
type UUIDGenerator struct{}

var _ Generator = (*UUIDGenerator)(nil)

// generates a new UUID identifier.
func (u *UUIDGenerator) Generate() string {
	return uuid.NewString()
}

// returns a new context with the UUID generator.
func (u *UUIDGenerator) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, u)
}
