package logging

func gray(s string) string {
	return "\033[38;05;240m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[38;05;197m" + s + "\033[0m"
}

func darkRed(s string) string {
	return "\033[38;05;124m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[38;05;215m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[38;05;69m" + s + "\033[0m"
}
