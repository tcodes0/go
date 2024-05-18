package reflectutil

import (
	"errors"
	"reflect"
)

// Merge combines the old struct with the new struct, iterating fields
// and picking the new field's value if non-zero and not ignored.
// Returns an updated copy of the old struct.
func Merge[T any](old, new *T, ignore []string) (copy *T, ignored []string, err error) {
	defer func() {
		if msg := recover(); msg != nil {
			copy = nil
			err = errors.New("merge: " + (msg).(string))
		}
	}()

	if old == nil {
		return new, nil, nil
	}

	c := Clone(*old)
	copy = &c

	if new == nil {
		return
	}

	vCopy := reflect.ValueOf(copy)
	vNew := reflect.ValueOf(new)

	if vNew.Kind() != reflect.Pointer || vCopy.Kind() != reflect.Pointer {
		return nil, nil, errors.New("merge: args must be pointers")
	}

	elemCopy := reflect.ValueOf(copy).Elem()
	elemNew := reflect.ValueOf(new).Elem()

	if elemNew.Kind() != reflect.Struct || elemCopy.Kind() != reflect.Struct {
		return nil, nil, errors.New("merge: args must be structs")
	}

	if elemNew.NumField() != elemCopy.NumField() {
		return nil, nil, errors.New("merge: args must have the same number of fields")
	}

fields:
	for i := 0; i < elemNew.NumField(); i++ {
		fCopy := elemCopy.Field(i)
		fNew := elemNew.Field(i)
		fName := elemNew.Type().Field(i).Name

		for _, g := range ignore {
			if g == fName {
				if !IsNilOrZero(fNew) {
					ignored = append(ignored, g)
				}

				continue fields
			}
		}

		if !IsNilOrZero(fNew) && fCopy.CanSet() {
			fCopy.Set(fNew)
		}
	}

	return copy, ignored, nil
}

func IsNilOrZero(v reflect.Value) bool {
	var nilKinds = []reflect.Kind{reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice}

	for _, nilKind := range nilKinds {
		if v.Kind() == nilKind {
			return v.IsNil()
		}
	}

	return v.IsZero()
}

// Clones the target, panicking if it is a pointer.
func Clone[T any](target T) T {
	if reflect.ValueOf(target).Kind() == reflect.Pointer {
		panic("value to clone must not be a pointer")
	}

	return target
}
