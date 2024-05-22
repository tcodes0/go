package logging

var (
	colorEnd  = "\033[0m"
	colorGray = "\033[38;05;240m"
)

func gray(s string) string {
	return colorGray + s
}

func lightGray(s string) string {
	return "\033[38;05;100m" + s + colorEnd
}

func brightRed(s string) string {
	// gray is added to color the log line information
	return "\033[38;05;197m" + s + colorEnd + colorGray
}

func red(s string) string {
	return "\033[38;05;124m" + s + colorEnd + colorGray
}

func yellow(s string) string {
	return "\033[38;05;215m" + s + colorEnd + colorGray
}

func blue(s string) string {
	return "\033[38;05;69m" + s + colorEnd + colorGray
}
