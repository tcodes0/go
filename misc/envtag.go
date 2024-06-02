package misc

import (
	"errors"
	"os"
	"reflect"
)

var (
	ErrNotString     = errors.New("only strings are supported")
	ErrNotAddresable = errors.New("field is not addressable")
)

// resolves a field's value to an env variable using a tag.
type EnvTag struct {
	Tag     string
	Default string
}

var _ FieldUpdater = (*EnvTag)(nil)

// updates a field with an env variable.
func (envTag EnvTag) UpdateField(field *reflect.StructField, valField reflect.Value) error {
	tag := field.Tag.Get(envTag.Tag)
	def := field.Tag.Get(envTag.Default)

	if tag == "" {
		return nil
	}

	if valField.Type() != reflect.TypeOf("") {
		return Wrap(ErrNotString, valField.String())
	}

	if valField.String() != "" {
		// do not overwrite fields
		return nil
	}

	tagValue := os.Getenv(tag)

	if tagValue == "" && def != "" {
		tagValue = def
	}

	if !valField.CanSet() {
		return ErrNotAddresable
	}

	valKey := reflect.ValueOf(tagValue)
	valField.Set(valKey)

	return nil
}
