package test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/httpflush"
)

func Test_MaxSize_Flushes(t *testing.T) {
	assert := require.New(t)
	writer := httpflush.NewMockResponseWriter(t)
	maxSize := httpflush.MaxSize{
		Max:    10,
		Writer: writer,
	}

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(5, nil).Once()

	n, err := maxSize.Write([]byte(""))
	assert.Equal(5, n)
	assert.NoError(err)

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(10, nil).Once()
	writer.On("Flush").Return(nil).Once()

	n, err = maxSize.Write([]byte(""))
	assert.Equal(10, n)
	assert.NoError(err)

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(10, nil).Once()
	writer.On("Flush").Return(nil).Once()

	n, err = maxSize.Write([]byte(""))
	assert.Equal(10, n)
	assert.NoError(err)
}
