package misc

import (
	"os"
	"reflect"
	"strconv"
)

type lookupEnv interface {
	string | int | bool
}

func LookupEnv[T lookupEnv](key string, fallback T) T {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	ref := reflect.ValueOf(fallback)

	if !ref.IsValid() {
		return fallback
	}

	//nolint:exhaustive // generic constraint
	switch ref.Kind() {
	default:
		return fallback
	case reflect.String:
		//nolint:forcetypeassert // generic constraint
		return reflect.ValueOf(val).Interface().(T)
	case reflect.Int:
		i, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		//nolint:forcetypeassert // generic constraint
		return reflect.ValueOf(i).Interface().(T)
	case reflect.Bool:
		bol, err := strconv.ParseBool(val)
		if err != nil {
			panic(err)
		}

		//nolint:forcetypeassert // generic constraint
		return reflect.ValueOf(bol).Interface().(T)
	}
}
