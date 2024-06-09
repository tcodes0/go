// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrNotStructPointer = errors.New("expected a struct pointer")

// FieldUpdater updates a field in a struct.
type FieldUpdater interface {
	UpdateField(field *reflect.StructField, base reflect.Value) error
}

// Applies the function to all fields in base struct.
func ApplyToFields[T any](updater FieldUpdater, base *T) (err error) {
	defer func() {
		if msg := recover(); msg != nil {
			err = fmt.Errorf("panic: %v", msg)
		}
	}()

	valBase := reflect.ValueOf(base)
	elemBase := valBase.Elem()

	if valBase.Kind() != reflect.Ptr || valBase.Elem().Kind() != reflect.Struct {
		return ErrNotStructPointer
	}

	typeBase := elemBase.Type()
	for i := range elemBase.NumField() {
		f := typeBase.Field(i)

		err := updater.UpdateField(&f, elemBase.Field(i))
		if err != nil {
			return Wrapf(err, "resolving field %s", typeBase.Field(i).Name)
		}
	}

	return nil
}
