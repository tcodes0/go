package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/configkey"
	"github.com/tcodes0/go/src/misc"
)

func TestEnvTag(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Test  string `key:"TEST"`
		Empty string
	}

	cfg := &config{}
	getKey := func(string) string {
		return "testing"
	}

	err := configkey.Run(cfg, getKey)
	assert.NoError(err)

	assert.Equal("testing", cfg.Test)
	assert.Equal("", cfg.Empty)
}

func TestEnvTag_Default(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Test string `key:"TEST" fallback:"foo"`
	}

	cfg := &config{}
	getKey := func(string) string {
		return ""
	}

	err := configkey.Run(cfg, getKey)
	assert.NoError(err)

	assert.Equal("foo", cfg.Test)
}

func TestEnvTag_Empty(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Test string `key:"TEST"`
	}

	cfg := &config{}
	getKey := func(string) string {
		return ""
	}

	err := configkey.Run(cfg, getKey)
	assert.Error(err)
	assert.ErrorIs(err, configkey.ErrEmpty)
}

func TestEnvTag_EmptyFallback(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Test string `key:"TEST" fallback:"-"`
	}

	cfg := &config{}
	getKey := func(string) string {
		return ""
	}

	err := configkey.Run(cfg, getKey)
	assert.NoError(err)
	assert.Equal("", cfg.Test)
}

func TestEnvTag_NoOverwrite(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Test string `key:"TEST"`
	}

	cfg := &config{
		Test: "testing",
	}
	getKey := func(string) string {
		return "foo"
	}

	err := configkey.Run(cfg, getKey)
	assert.NoError(err)
	assert.Equal("testing", cfg.Test)
}

func TestEnvTag_ErrNonString(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Ptr  *string
		Sl   []byte
		Ok   bool
		Test bool `key:"TEST"`
	}

	cfg := &config{
		Ok:   true,
		Ptr:  misc.PointerTo("foo"),
		Test: true,
		Sl:   []byte("foo"),
	}
	getKey := func(string) string {
		return "foo"
	}

	err := configkey.Run(cfg, getKey)
	assert.Error(err)
	assert.ErrorIs(err, configkey.ErrNonString)
}

func TestEnvTag_Noop(t *testing.T) {
	assert := require.New(t)

	type config struct {
		Ptr *string
		Sl  []byte
		Ok  bool
	}

	cfg := &config{
		Ok:  true,
		Ptr: misc.PointerTo("foo"),
		Sl:  []byte("foo"),
	}
	getKey := func(string) string {
		return "foo"
	}
	expected := &config{
		Ok:  true,
		Ptr: misc.PointerTo("foo"),
		Sl:  []byte("foo"),
	}

	err := configkey.Run(cfg, getKey)
	assert.NoError(err)
	assert.Equal(expected, cfg)
}
