package reflectutil

import "reflect"

var nilKinds = []reflect.Kind{
	reflect.Chan,
	reflect.Func,
	reflect.Interface,
	reflect.Map,
	reflect.Pointer,
	reflect.Slice,
}

func IsNil(value reflect.Value) bool {
	for _, nk := range nilKinds {
		if value.Kind() == nk {
			return value.IsNil()
		}
	}

	return false
}

func IsZero(value reflect.Value) bool {
	for _, nk := range nilKinds {
		if value.Kind() != nk {
			return value.IsZero()
		}
	}

	return false
}
