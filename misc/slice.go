// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

// Find returns the first item in the set that satisfies the finder function.
func Find[T any](set []T, finder func(item T) bool) (*T, bool) {
	for _, t := range set {
		if finder(t) {
			return &t, true
		}
	}

	return nil, false
}

// Uniq returns a new slice containing no duplicates.
func Uniq[T comparable](set []T) []T {
	seen := make(map[T]bool)
	unique := make([]T, 0, len(set))

	for _, t := range set {
		if _, exists := seen[t]; !exists {
			seen[t] = true

			unique = append(unique, t)
		}
	}

	return unique
}
