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
