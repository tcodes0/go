// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/samber/lo"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var (
	ignore      = regexp.MustCompile(`test$|\.local.*|cmd/template`)
	globs       = []string{"*/*.go", "*/*/*.go"}
	EnvColor    = "T0_COLOR"
	EnvLogLevel = "T0_LOGLEVEL"
)

func FindPackages(logger *logging.Logger) ([]string, error) {
	goFiles := make([]string, 0)

	for _, glob := range globs {
		g, err := filepath.Glob(glob)
		if err != nil {
			return nil, misc.Wrapfl(err)
		}

		goFiles = append(goFiles, g...)
	}

	dirs := make([]string, 0, len(goFiles))

	for _, file := range goFiles {
		dirs = append(dirs, path.Dir(file))
	}

	dirs = lo.Uniq(dirs)
	out := make([]string, 0, len(dirs))

	for _, module := range dirs {
		if ignore.MatchString(module) {
			logger.Debugf("ignored %s", module)

			continue
		}

		out = append(out, module)
	}

	slices.Sort(out)

	return out, nil
}

// Overwrites the file, does not create the file.
func OverwriteFile(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0)
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
	format := `environment variables:
- %s     toggle logger colored output (default: false)
- %s  1 - 5, 1 is debug. The higher the less logs (default: 2)`

	return fmt.Sprintf(format, EnvColor, EnvLogLevel)
}
