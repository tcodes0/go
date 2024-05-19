package test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/reflectutil"
)

func TestApplyFieldResolver(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	type config struct {
		Foo string
		Bar string
	}

	cfg := &config{}
	resolver := reflectutil.NewMockFieldResolver(t)
	resolver.On("Resolve", mock.AnythingOfType("*reflect.StructField"), mock.AnythingOfType("reflect.Value")).
		Return(nil).
		Twice()

	err := reflectutil.ApplyFieldResolver(resolver, cfg)
	assert.NoError(err)
}

// func TestEnvTag_Default(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Test string `fallback:"foo" key:"TEST"`
// 	}

// 	cfg := &config{}
// 	getKey := func(string) string {
// 		return ""
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.NoError(err)

// 	assert.Equal(foo, cfg.Test)
// }

// func TestEnvTag_Empty(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Test string `key:"TEST"`
// 	}

// 	cfg := &config{}
// 	getKey := func(string) string {
// 		return ""
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.Error(err)
// 	assert.ErrorIs(err, configkey.ErrEmpty)
// }

// func TestEnvTag_EmptyFallback(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Test string `fallback:"-" key:"TEST"`
// 	}

// 	cfg := &config{}
// 	getKey := func(string) string {
// 		return ""
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.NoError(err)
// 	assert.Equal("", cfg.Test)
// }

// func TestEnvTag_NoOverwrite(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Test string `key:"TEST"`
// 	}

// 	cfg := &config{
// 		Test: "testing",
// 	}
// 	getKey := func(string) string {
// 		return foo
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.NoError(err)
// 	assert.Equal("testing", cfg.Test)
// }

// func TestEnvTag_ErrNonString(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Ptr  *string
// 		Sl   []byte
// 		Ok   bool
// 		Test bool `key:"TEST"`
// 	}

// 	cfg := &config{
// 		Ok:   true,
// 		Ptr:  misc.PointerTo(foo),
// 		Test: true,
// 		Sl:   []byte(foo),
// 	}
// 	getKey := func(string) string {
// 		return foo
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.Error(err)
// 	assert.ErrorIs(err, configkey.ErrNonString)
// }

// func TestEnvTag_Noop(t *testing.T) {
// 	t.Parallel()
// 	assert := require.New(t)

// 	type config struct {
// 		Ptr *string
// 		Sl  []byte
// 		Ok  bool
// 	}

// 	cfg := &config{
// 		Ok:  true,
// 		Ptr: misc.PointerTo(foo),
// 		Sl:  []byte(foo),
// 	}
// 	getKey := func(string) string {
// 		return foo
// 	}
// 	expected := &config{
// 		Ok:  true,
// 		Ptr: misc.PointerTo(foo),
// 		Sl:  []byte(foo),
// 	}

// 	err := configkey.Run(cfg, getKey)
// 	assert.NoError(err)
// 	assert.Equal(expected, cfg)
// }
