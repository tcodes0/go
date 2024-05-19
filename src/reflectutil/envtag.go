package reflectutil

import (
	"errors"
	"os"
	"reflect"

	"github.com/tcodes0/go/src/errutil"
)

var ErrNotString = errors.New("only strings are supported")

// Resolves a field's value to an env variable using a tag
type EnvTag struct {
	Tag     string
	Default string
}

var _ FieldResolver = (*EnvTag)(nil)

func (resolver EnvTag) Resolve(field *reflect.StructField, valField reflect.Value) error {
	tag := field.Tag.Get(resolver.Tag)
	def := field.Tag.Get(resolver.Default)

	if tag == "" {
		return nil
	}

	if valField.Type() != reflect.TypeOf("") {
		return errutil.Wrap(ErrNotString, valField.String())
	}

	if valField.String() != "" {
		// do not overwrite fields
		return nil
	}

	tagValue := os.Getenv(tag)

	if tagValue == "" && def != "" {
		tagValue = def
	}

	valKey := reflect.ValueOf(tagValue)
	valField.Set(valKey)

	return nil
}
