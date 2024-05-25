package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/identifier"
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
