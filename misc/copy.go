// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

// returns a pointer copy of value.
func ToPtr[T any](x T) *T {
	return &x
}

// returns a copy of the value ptr points to.
func CopyPointed[T any](ptr *T) T {
	if ptr == nil {
		z := new(T)

		return *z
	}

	c := Copy(*ptr)

	return c
}

// return a copy of T.
func Copy[T any](val T) T {
	return val
}
