// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package identifier

import (
	"context"
	"strconv"
)

type StaticGenerator struct {
	Prefix string
	Count  int
}

var _ Generator = (*StaticGenerator)(nil)

// generates a new static identifier.
func (static *StaticGenerator) Generate() string {
	id := static.Prefix + "-" + strconv.Itoa(static.Count)
	static.Count += 1

	return id
}

// returns a new context with the static generator.
func (static *StaticGenerator) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, static)
}
