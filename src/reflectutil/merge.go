package reflectutil

import (
	"errors"
	"reflect"
)

var (
	ErrArgPointer = errors.New("merge: args must be pointers")
	ErrArgStruct  = errors.New("merge: args must be structs")
	ErrSameFields = errors.New("merge: args must have the same number of fields")
)

// Merge combines the base struct with the actual struct, iterating fields
// and picking the actual field's value if non-zero and not ignored.
// Returns an updated copy of the base struct.
func Merge[T any](base, actual *T, ignore []string) (clone *T, ignored []string, err error) {
	defer func() {
		if msg := recover(); msg != nil {
			clone = nil
			msgs, ok := msg.(string)

			if ok {
				err = errors.New("merge: " + msgs)
			}
		}
	}()

	if base == nil {
		return actual, nil, nil
	}

	c := CopyOf(*base)
	clone = &c

	if actual == nil {
		return clone, nil, nil
	}

	valClone := reflect.ValueOf(clone)
	valActual := reflect.ValueOf(actual)

	valClone, valActual, err = mergeErrs(valClone, valActual)
	if err != nil {
		return nil, nil, err
	}

fields:
	for i := range valActual.NumField() {
		fieldCopy := valClone.Field(i)
		fieldNew := valActual.Field(i)
		fieldName := valActual.Type().Field(i).Name
		hasValue := !IsNil(fieldNew) || !IsZero(fieldNew)

		for _, g := range ignore {
			if g == fieldName {
				if hasValue {
					ignored = append(ignored, g)
				}

				continue fields
			}
		}

		if hasValue && fieldCopy.CanSet() {
			fieldCopy.Set(fieldNew)
		}
	}

	return clone, ignored, nil
}

func mergeErrs(base, actual reflect.Value) (baseElem, actualElem reflect.Value, err error) {
	if actual.Kind() != reflect.Pointer || base.Kind() != reflect.Pointer {
		return reflect.Value{}, reflect.Value{}, ErrArgPointer
	}

	base = reflect.ValueOf(base).Elem()
	actual = reflect.ValueOf(actual).Elem()

	if actual.Kind() != reflect.Struct || base.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.Value{}, ErrArgStruct
	}

	if actual.NumField() != base.NumField() {
		return reflect.Value{}, reflect.Value{}, ErrSameFields
	}

	return base, actual, nil
}
