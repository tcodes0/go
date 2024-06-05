// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

import "time"

// returns a duration of i.
func Seconds(i int) time.Duration {
	return time.Duration(i) * time.Second
}

// returns a duration of i.
func Minutes(i int) time.Duration {
	return time.Duration(i) * time.Minute
}

// returns a duration of i.
func Hours(i int) time.Duration {
	return time.Duration(i) * time.Hour
}

// returns a duration of i.
func Days(i int) time.Duration {
	return time.Duration(i) * 24 * time.Hour
}
