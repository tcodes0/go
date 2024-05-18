package configkey

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotStructPointer = errors.New("expected a pointer to a struct")
	ErrEmpty            = errors.New("empty value without fallback")
	ErrNonString        = errors.New("non string tagged")

	tagKey      string = "key"
	tagFallback string = "fallback"
	optional    string = "-"
)

// Sets (only!) string values in the struct pointed to, using getKey for values and
// extracting keys from struct tags. Empty values with tags are not allowed
// without a fallback value. Use '-' as fallback to allow empty values. Won't overwrite
// a field that is already set.
func Run(structPtr any, getKey func(key string) string) (err error) {
	defer func() {
		if msg := recover(); msg != nil {
			err = fmt.Errorf("panic: %v", msg)
		}
	}()

	rValue := reflect.ValueOf(structPtr)
	if rValue.Kind() != reflect.Ptr {
		return ErrNotStructPointer
	}

	rValue = rValue.Elem()
	if rValue.Kind() != reflect.Struct {
		return ErrNotStructPointer
	}

	rType := rValue.Type()

	for i := 0; i < rType.NumField(); i++ {
		rStructFieldValue := rValue.Field(i)
		rStructField := rType.Field(i)

		envKey := rStructField.Tag.Get(tagKey)
		fallbackValue := rStructField.Tag.Get(tagFallback)

		if envKey == "" {
			// ignore fields without a tagKey
			continue
		}

		if rStructFieldValue.Type() != reflect.TypeOf("") {
			return fmt.Errorf("%s: %w", rStructFieldValue.String(), ErrNonString)
		}

		if rStructFieldValue.String() != "" {
			// field is a string and is already set
			continue
		}

		gotValue := getKey(envKey)

		if gotValue == "" && fallbackValue != "" && fallbackValue != optional {
			gotValue = fallbackValue
		}

		if gotValue == "" && fallbackValue != optional {
			return fmt.Errorf("key %s: %w", envKey, ErrEmpty)
		}

		rEnvValue := reflect.ValueOf(gotValue)
		rStructFieldValue.Set(rEnvValue)
	}

	return err
}
