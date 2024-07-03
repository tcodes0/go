// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package hue

import "fmt"

var (
	// end terminal formats and colors.
	End       = "\033[0m"
	Gray      = 240
	Brown     = 100
	BrightRed = 197
	Red       = 124
	Yellow    = 215
	Blue      = 69
)

// terminal ansi escape code for color.
func TermColor(n int) string {
	return fmt.Sprintf("\033[38;05;%dm", n)
}

// print in color, does not terminate color; append the exported var.
func Printc(color int, messages ...string) string {
	out := TermColor(color)
	for _, s := range messages {
		out += s
	}

	return out
}
