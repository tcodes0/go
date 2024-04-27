package test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/httpflush"
)

func Test_MaxSize_NoFlushSmaller(t *testing.T) {
	assert := require.New(t)
	writer := httpflush.NewMockresponseWriter(t)
	maxSize := httpflush.MaxSize{
		Max:    10,
		Writer: writer,
	}

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(5, nil).Once()

	n, err := maxSize.Write([]byte(""))
	assert.Equal(4, n)
	assert.NoError(err)
}

func TestMaxSize_FlushesMany(t *testing.T) {
	assert := require.New(t)
	writer := httpflush.NewMockresponseWriter(t)
	maxSize := httpflush.MaxSize{
		Max:    10,
		Writer: writer,
	}

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(20, nil).Once()
	writer.On("Flush").Return(nil).Once()

	n, err := maxSize.Write([]byte(""))
	assert.Equal(20, n)
	assert.NoError(err)

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(11, nil).Once()
	writer.On("Flush").Return(nil).Once()

	n, err = maxSize.Write([]byte(""))
	assert.Equal(11, n)
	assert.NoError(err)
}
