package reflectutil

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tcodes0/go/src/misc"
)

var (
	ErrArgPointer = errors.New("merge: args must be pointers")
	ErrArgStruct  = errors.New("merge: args must be structs")
	ErrSameFields = errors.New("merge: args must have the same number of fields")
)

// Combines the base struct with the partial struct, iterating fields
// and picking the partial field's value if not zero and not ignored.
// Returns an updated copy of the base struct.
func Merge[T any](base, partial *T, ignore []string) (clone *T, ignored []string, err error) {
	defer func() {
		if x := recover(); x != nil {
			clone = nil
			err = fmt.Errorf("merge panic: %#v", x)
		}
	}()

	if base == nil {
		return partial, nil, nil
	}

	c := misc.Copy(*base)
	clone = &c

	if partial == nil {
		return clone, nil, nil
	}

	valClone := reflect.ValueOf(clone)
	valPartial := reflect.ValueOf(partial)

	valClone, valPartial, err = mergeErrs(valClone, valPartial)
	if err != nil {
		return nil, nil, err
	}

fields:
	for i := range valPartial.NumField() {
		fieldCopy := valClone.Field(i)
		fieldNew := valPartial.Field(i)
		fieldName := valPartial.Type().Field(i).Name
		hasValue := !IsNil(fieldNew) && !IsZero(fieldNew)

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

func mergeErrs(base, partial reflect.Value) (baseElem, partialElem reflect.Value, err error) {
	if partial.Kind() != reflect.Pointer || base.Kind() != reflect.Pointer {
		return reflect.Value{}, reflect.Value{}, ErrArgPointer
	}

	base = base.Elem()
	partial = partial.Elem()

	if partial.Kind() != reflect.Struct || base.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.Value{}, ErrArgStruct
	}

	if partial.NumField() != base.NumField() {
		return reflect.Value{}, reflect.Value{}, ErrSameFields
	}

	return base, partial, nil
}
