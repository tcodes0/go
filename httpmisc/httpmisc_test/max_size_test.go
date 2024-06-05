// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package httpmisc_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcodes0/go/httpmisc"
)

func TestMaxSizeWrite(t *testing.T) {
	t.Parallel()
	writer1 := httpmisc.NewMockwriterFlusher(t)
	writer2 := httpmisc.NewMockwriterFlusher(t)
	writer3 := httpmisc.NewMockwriterFlusher(t)

	writer1.Expect().Write([]uint8{}).Return(5, nil).Once()
	writer2.Expect().Write([]uint8{}).Return(10, nil).Once()
	writer3.Expect().Write([]uint8{}).Return(20, nil).Once()
	writer3.Expect().Flush()

	tests := []struct {
		maxSize *httpmisc.MaxSize
		name    string
		wantN   int
		wantErr bool
	}{
		{
			name:    "no flush smaller",
			maxSize: &httpmisc.MaxSize{Max: 10, Writer: writer1},
			wantN:   5,
			wantErr: false,
		},
		{
			name:    "no flush equal",
			maxSize: &httpmisc.MaxSize{Max: 10, Writer: writer2},
			wantN:   10,
			wantErr: false,
		},
		{
			name:    "flush larger",
			maxSize: &httpmisc.MaxSize{Max: 10, Writer: writer3},
			wantN:   20,
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gotN, err := test.maxSize.Write([]byte(""))
			if (err != nil) != test.wantErr {
				t.Errorf("MaxSize.Write() error = %v, wantErr %v", err, test.wantErr)

				return
			}

			if gotN != test.wantN {
				t.Errorf("MaxSize.Write() = %v, want %v", gotN, test.wantN)
			}
		})
	}
}

func TestMaxSizeWrite_FlushesMany(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	writer := httpmisc.NewMockwriterFlusher(t)
	maxSize := httpmisc.MaxSize{
		Max:    10,
		Writer: writer,
	}

	writer.On("Write", []uint8{}).Return(20, nil).Once()
	writer.On("Flush").Return(nil).Once()

	writtenBytes, err := maxSize.Write([]byte(""))
	assert.Equal(20, writtenBytes)
	assert.NoError(err)

	writer.On("Write", []uint8{}).Return(11, nil).Once()
	writer.On("Flush").Return(nil).Once()

	writtenBytes, err = maxSize.Write([]byte(""))
	assert.Equal(11, writtenBytes)
	assert.NoError(err)
}
