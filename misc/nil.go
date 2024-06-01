package misc

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

	var val reflect.Value

	switch tVal := value.(type) {
	case reflect.Value:
		if !tVal.IsValid() {
			return true
		}

		val = tVal
	default:
		val = reflect.ValueOf(value)
	}

	for _, nk := range nilKinds {
		if val.Kind() == nk {
			return val.IsNil()
		}
	}

	return false
}

// wraps reflect.Value{}.IsZero() but returns false if would panic.
func IsZero(value any) bool {
	var val reflect.Value

	switch tVal := value.(type) {
	case reflect.Value:
		if !tVal.IsValid() {
			return true
		}

		val = tVal
	default:
		val = reflect.ValueOf(value)
	}

	return val.IsValid() && val.IsZero()
}
