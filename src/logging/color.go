package logging

func green(s string) string {
	return "\033[38;05;029m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[38;05;197m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[38;05;215m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[38;05;69m" + s + "\033[0m"
}
