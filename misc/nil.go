// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

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
