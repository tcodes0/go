package misc

import "time"

func Seconds(i int) time.Duration {
	return time.Duration(i) * time.Second
}

func Minutes(i int) time.Duration {
	return time.Duration(i) * time.Minute
}

func Hours(i int) time.Duration {
	return time.Duration(i) * time.Hour
}

func Days(i int) time.Duration {
	return time.Duration(i) * 24 * time.Hour
}
