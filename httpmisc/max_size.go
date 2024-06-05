// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package httpmisc

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrWriterNotFlusher = errors.New("http.ResponseWriter does not implement http.Flusher")

// MaxSize wraps an http.ResponseWriter and flushes every time more than max bytes are written.
// If max is smaller than the first write, it will never flush.
type MaxSize struct {
	Writer                http.ResponseWriter
	Max                   int
	writtenSinceLastFlush int
}

var _ writerFlusher = (*MaxSize)(nil)

// implementation of http.ResponseWriter.Write.
func (maxSize *MaxSize) Write(b []byte) (n int, err error) {
	n, err = maxSize.Writer.Write(b)
	if err != nil {
		return n, fmt.Errorf("%s: %w", "maxSize.Writer.Write", err)
	}

	maxSize.writtenSinceLastFlush += n

	if maxSize.writtenSinceLastFlush > maxSize.Max {
		_, ok := maxSize.Writer.(http.Flusher)
		if !ok {
			return n, ErrWriterNotFlusher
		}

		maxSize.Flush()
	}

	return n, nil
}

// implementation of http.ResponseWriter.Header.
func (maxSize MaxSize) Header() http.Header {
	return maxSize.Writer.Header()
}

// implementation of http.ResponseWriter.WriteHeader.
func (maxSize MaxSize) WriteHeader(statusCode int) {
	maxSize.Writer.WriteHeader(statusCode)
}

// Flush flushes the writer.
func (maxSize *MaxSize) Flush() {
	f, _ := maxSize.Writer.(http.Flusher)
	f.Flush()

	maxSize.writtenSinceLastFlush = 0
}
