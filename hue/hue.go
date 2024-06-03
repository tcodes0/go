package hue

import "fmt"

var (
	// end terminal formats and colors.
	TermEnd   = "\033[0m"
	Gray      = 240
	Brown     = 100
	BrightRed = 197
	Red       = 124
	Yellow    = 215
	Blue      = 69
)

// terminal escape code for color.
func TermColor(n int) string {
	return fmt.Sprintf("\033[38;05;%dm", n)
}

// color print, does not terminate color; append the exported var.
func Cprint(color int, messages ...string) string {
	out := TermColor(color)
	for _, s := range messages {
		out += s
	}

	return out
}
