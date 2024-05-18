package reflectutil

// returns a copy of the value ptr points to.
func CopyPointed[T any](ptr *T) T {
	v := *ptr
	c := CopyOf(v)

	return c
}

// return a copy of the value.
func CopyOf[T any](val T) T {
	return val
}
