// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package identifier_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/identifier"
)

func TestStaticGenerator_Generate(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	static := &identifier.StaticGenerator{
		Prefix: "products",
		Count:  0,
	}

	assert.Equal("products-0", static.Generate())
	assert.Equal("products-1", static.Generate())
}
