package reflectutil

import "reflect"

func IsNil(value reflect.Value) bool {
	nilKinds := []reflect.Kind{
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice,
	}

	for _, nk := range nilKinds {
		if value.Kind() == nk {
			return value.IsNil()
		}
	}

	return false
}

func IsZero(value reflect.Value) bool {
	return value.IsValid() && value.IsZero()
}
