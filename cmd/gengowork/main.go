// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tcodes0/go/logging"
)

var flagset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

func main() {
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, the higher the less logs.")
	fColor := flagset.Bool("color", false, "colored logging output. (default false)")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)

	err = genGoWork(*logger)
	if err != nil {
		logger.Fatalf("fatal: %v", err)
	}
}

func usageExit(err error) {
	fmt.Println()
	fmt.Println("generates go.work file")
	fmt.Println()

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Printf("error: %v\n", err)
	}

	os.Exit(1)
}

func genGoWork(_ logging.Logger) error {
	return nil
}
