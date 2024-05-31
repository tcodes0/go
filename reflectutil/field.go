package reflectutil

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tcodes0/go/errutil"
)

var ErrNotStructPointer = errors.New("expected a struct pointer")

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
			return errutil.Wrapf(err, "resolving field %s", typeBase.Field(i).Name)
		}
	}

	return nil
}
