package misc

// returns a pointer copy of value.
func ToPtr[T any](x T) *T {
	return &x
}

// returns a copy of the value ptr points to.
func CopyPointed[T any](ptr *T) T {
	if ptr == nil {
		z := new(T)

		return *z
	}

	c := Copy(*ptr)

	return c
}

// return a copy of T.
func Copy[T any](val T) T {
	return val
}
