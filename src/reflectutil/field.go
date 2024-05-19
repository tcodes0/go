package reflectutil

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tcodes0/go/src/errutil"
)

var ErrNotStructPointer = errors.New("expected a pointer to a struct")

type FieldResolver interface {
	Resolve(field *reflect.StructField, base reflect.Value) error
}

// Applies a FieldResolver to all fields in a struct.
func ApplyFieldResolver[T any](fResolver FieldResolver, base *T) (err error) {
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

		err := fResolver.Resolve(&f, elemBase.Field(i))
		if err != nil {
			return errutil.Wrapf(err, "resolving field %s", typeBase.Field(i).Name)
		}
	}

	return nil
}
