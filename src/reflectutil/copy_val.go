package reflectutil

func CopyPointerToValue[T any](ptr *T) T {
	v := *ptr
	c := CopyValue(v)
	return c
}

func CopyValue[T any](val T) T {
	return val
}
