// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
)

// Wrap wraps an error with a message.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return Wrap(err, fmt.Sprintf(format, args...))
}

// Wrapfl wraps an error with file:line information.
func Wrapfl(err error) error {
	if err == nil {
		return nil
	}

	file, line := "?", "?"

	_, f, l, ok := runtime.Caller(1)
	if ok {
		file, line = filepath.Base(f), strconv.Itoa(l)
	}

	return fmt.Errorf("%s:%s: %w", file, line, err)
}
