package reflectutil_test

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
	updater := reflectutil.NewMockFieldUpdater(t)
	updater.Expect().UpdateField(mock.AnythingOfType("*reflect.StructField"), mock.AnythingOfType("reflect.Value")).
		Return(nil).
		Twice()

	err := reflectutil.ApplyToFields(updater, cfg)
	assert.NoError(err)
}
