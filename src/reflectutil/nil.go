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

// wraps reflect.Value{}.IsNil() but returns false if would panic.
func IsNil(value any) bool {
	if value == nil {
		return true
	}

	r := reflect.ValueOf(value)

	for _, nk := range nilKinds {
		if r.Kind() == nk {
			return r.IsNil()
		}
	}

	return false
}

// wraps reflect.Value{}.IsZero() but returns false if would panic.
func IsZero(value any) bool {
	r := reflect.ValueOf(value)

	return r.IsValid() && r.IsZero()
}
