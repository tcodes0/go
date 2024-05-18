package misc

// returns a pointer copy of value.
func PointerTo[T any](x T) *T {
	return &x
}
