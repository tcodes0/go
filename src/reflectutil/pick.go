package reflectutil

import "reflect"

// Returns the first non-nil or non-zero value from the list of values.
func PickNonZero[T any](values ...T) T {
	for _, v := range values {
		value := reflect.ValueOf(v)
		if !IsNil(value) || !IsZero(value) {
			return v
		}
	}

	var t T
	zeroT, _ := reflect.Zero(reflect.TypeOf(t)).Interface().(T)

	return zeroT
}

// Returns the default if value is zero.
func Default[T any](value, defaultVal T) T {
	return PickNonZero(value, defaultVal)
}
