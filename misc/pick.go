// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

// from left to right returns the first valid value or the last value
// if all are nil or zero.
func PickValid[T any](values ...T) T {
	if len(values) == 0 {
		return *new(T)
	}

	for _, val := range values {
		if !IsNil(val) && !IsZero(val) {
			return val
		}
	}

	return values[len(values)-1]
}

// returns the default if value is nil or zero.
func Default[T any](value, defaultVal T) T {
	return PickValid(value, defaultVal)
}
