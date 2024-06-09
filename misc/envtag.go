// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

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
