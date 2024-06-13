// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var (
	ignore      = regexp.MustCompile(`test$|\.local.*|cmd/template|^cmd$`)
	EnvColor    = "CMD_COLOR"
	EnvLogLevel = "CMD_LOGLEVEL"
)

func FindModules(logger logging.Logger) ([]string, error) {
	cmd := exec.Command("find", ".", "-mindepth", "2", "-maxdepth", "3", "-type", "f", "-name", "*.go", "-exec", "dirname", "{}", ";")

	findOut, err := cmd.CombinedOutput()
	if err != nil {
		return nil, misc.Wrapf(err, "finding, %s", findOut)
	}

	logger.Debug().Logf("find output: %s", findOut)

	modules := strings.Split(string(findOut), "\n")
	modules = lo.Uniq(modules)

	out := make([]string, 0, len(modules))

	for _, module := range modules {
		module = strings.Replace(module, "./", "", 1)

		if ignore.MatchString(module) {
			logger.Debug().Logf("ignored %s", module)

			continue
		}

		if module == "" {
			continue
		}

		out = append(out, module)
	}

	slices.Sort(out)

	return out, nil
}

func WriteFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return misc.Wrap(err, "opening")
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return misc.Wrap(err, "stat")
	}

	if int64(len(data)) < stat.Size() {
		// new file is smaller, truncate to new size
		err = file.Truncate(int64(len(data)))
		if err != nil {
			return misc.Wrap(err, "truncating")
		}
	}

	_, err = file.Write(data)
	if err != nil {
		return misc.Wrap(err, "writing")
	}

	return nil
}

func EnvVarUsage() string {
	format := "environment variables:\n"
	format += "%s\t toggle logger colored output (default: false)\n"
	format += "%s\t set log level, 1 - 5, 1 is debug. The higher the less logs (default: 2)"

	return fmt.Sprintf(format, EnvColor, EnvLogLevel)
}
