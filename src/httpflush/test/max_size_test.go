package test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/src/httpflush"
)

func TestMaxSize_Write(t *testing.T) {
	writer1 := httpflush.NewMockresponseWriter(t)
	writer2 := httpflush.NewMockresponseWriter(t)
	writer3 := httpflush.NewMockresponseWriter(t)

	writer1.On("Write", mock.AnythingOfType("[]uint8")).Return(5, nil).Once()
	writer2.On("Write", mock.AnythingOfType("[]uint8")).Return(10, nil).Once()
	writer3.On("Write", mock.AnythingOfType("[]uint8")).Return(20, nil).Once()
	writer3.On("Flush").Return(nil).Once()

	tests := []struct {
		name    string
		maxSize *httpflush.MaxSize
		wantN   int
		wantErr bool
	}{
		{
			name:    "no flush smaller",
			maxSize: &httpflush.MaxSize{Max: 10, Writer: writer1},
			wantN:   5,
			wantErr: false,
		},
		{
			name:    "no flush equal",
			maxSize: &httpflush.MaxSize{Max: 10, Writer: writer2},
			wantN:   10,
			wantErr: false,
		},
		{
			name:    "flush larger",
			maxSize: &httpflush.MaxSize{Max: 10, Writer: writer3},
			wantN:   20,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, err := tt.maxSize.Write([]byte(""))
			if (err != nil) != tt.wantErr {
				t.Errorf("MaxSize.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("MaxSize.Write() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
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
