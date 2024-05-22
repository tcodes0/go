package hue

import "fmt"

var (
	TermEnd   = "\033[0m"
	Gray      = 240
	Brown     = 100
	BrightRed = 197
	Red       = 124
	Yellow    = 215
	Blue      = 69
)

func TermColor(n int) string {
	return fmt.Sprintf("\033[38;05;%dm", n)
}

func Cprint(color int, args ...string) string {
	out := TermColor(color)
	for _, s := range args {
		out += s
	}

	return out
}
