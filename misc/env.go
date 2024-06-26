// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package misc

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type lookupEnv interface {
	string | int | bool
}

// generic os.LookupEnv with fallback if value is not set or empty.
func LookupEnv[T lookupEnv](key string, fallback T) T {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
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

func DotEnv(path string, noisy bool) {
	file, err := os.Open(path)
	if err != nil {
		if noisy {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}

		return
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			if noisy {
				fmt.Fprint(os.Stderr, Wrap(err, "scanning .env file"))
			}

			return
		}

		key, val, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			if noisy {
				fmt.Fprintf(os.Stderr, "= not found in line: %s\n", scanner.Text())
			}

			continue
		}

		os.Setenv(key, val)
	}
}
