// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

const (
	actionLink   = "link"
	actionRemove = "remove"
	actionBak    = "backup"
)

var (
	logger       = &logging.Logger{}
	flagset      = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	actions      = []string{actionLink, actionRemove, actionBak}
	descriptions = []string{
		"create symbolic link; provide files in pairs, one per line, first is source, second is link",
		"delete; one file per line",
		"copy with .bak extension; one file per line",
	}
)

func main() {
	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Fatalf("%v", msg)
		}

		if err != nil {
			logger.Error().Logf("%v", err)
			os.Exit(1)
		}
	}()

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	fFiles := flagset.String("files", "", "path to newline separated list of files (required)")
	fAction := flagset.String("action", "", "action to take on files (required)")
	fCommitL := flagset.Bool("commit", false, "apply changes (default: false)")
	fCommitS := flagset.Bool("c", false, "apply changes (default: false)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		usageExit(err)
	}

	fDryrun := !(*fCommitL || *fCommitS)

	files, err := readConfig(*fFiles)
	if err != nil {
		usageExit(err)
	}

	if *fAction == "" {
		usageExit(errors.New("empty action"))
	}

	if _, found := lo.Find(actions, func(a string) bool { return a == *fAction }); !found {
		usageExit(fmt.Errorf("unknown action %s", *fAction))
	}

	files = envVarResolver(files)
	err = filer(files, *fAction, fDryrun)
}

func readConfig(configPath string) ([]string, error) {
	if configPath == "" {
		return nil, errors.New("empty config")
	}

	cfgFile, err := os.Open(configPath)
	if err != nil {
		return nil, misc.Wrap(err, "open")
	}

	raw, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, misc.Wrap(err, "read")
	}

	files := strings.Split(string(raw), "\n")
	files = lo.Filter(files, func(file string, _ int) bool {
		return file != "" && !strings.HasPrefix(file, "#")
	})

	if len(files) == 0 {
		return nil, errors.New("empty config")
	}

	return files, nil
}

func usageExit(err error) {
	flagset.Usage()
	fmt.Println("perform an action on a list of files.")
	fmt.Println("changes nothing by default, pass -commit")
	fmt.Println("comments with # and newlines are ignored in config file.")
	fmt.Println()
	fmt.Println("actions available:")

	for i, action := range actions {
		fmt.Printf("- %s: %s\n", action, descriptions[i])
	}

	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())

	if err != nil && !errors.Is(err, flag.ErrHelp) {
		logger.Error().Logf("%s", err.Error())
	}

	os.Exit(1)
}

func envVarResolver(files []string) []string {
	envs := map[string]string{}
	envs["$HOME"] = os.Getenv("HOME")
	out := make([]string, len(files))

	for i, file := range files {
		out[i] = strings.ReplaceAll(file, "$HOME", envs["$HOME"])
	}

	return out
}

func filer(files []string, action string, dryrun bool) error {
	switch action {
	case actionLink:
		if len(files)%2 != 0 {
			return fmt.Errorf("symlink: file count must be even, got %v", len(files))
		}

		for i, file := range files {
			if i%2 != 0 {
				continue
			}

			err := link(file, files[i+1], dryrun)
			if err != nil {
				return err
			}
		}
	case actionRemove:
		for _, file := range files {
			err := remove(file, dryrun)
			if err != nil {
				return err
			}
		}
	case actionBak:
		for _, file := range files {
			err := bak(file, dryrun)
			if err != nil {
				return err
			}
		}
	}

	if dryrun {
		fmt.Printf("to apply changes run: %s -commit", strings.Join(os.Args, " "))
	}

	return nil
}

func link(source, link string, dryrun bool) error {
	_, err := os.Stat(source)
	if err != nil {
		return misc.Wrapf(err, "stat")
	}

	linkDir := filepath.Dir(link)

	_, err = os.Stat(linkDir)
	if err != nil {
		logger.Logf("try: mkdir -p %s", linkDir)

		return misc.Wrapf(err, "stat dir")
	}

	// do not follow symlinks!
	lStat, err := os.Lstat(link)
	if err == nil {
		logger.Warn().Logf("skip: file exists %s", link)

		if lStat.Mode()&os.ModeSymlink != 0 {
			_, err = os.Stat(link)
			if err != nil {
				logger.Error().Logf("broken link %s", link)
				logger.Logf("try: unlink %s", link)
			}
		}

		return nil
	}

	if dryrun {
		_, err = fmt.Printf("link %s -> %s\n", link, source)
		if err != nil {
			logger.Error().Logf("println: %v", err)
		}

		return nil
	}

	err = os.Symlink(source, link)
	if err != nil {
		return misc.Wrapf(err, "symlink")
	}

	return nil
}

func remove(target string, dryrun bool) error {
	stat, err := os.Stat(target)
	if err != nil {
		logger.Warn().Logf("skip: file not found %s", target)

		//nolint:nilerr // func about removing files
		return nil
	}

	if stat.IsDir() {
		entries, e := os.ReadDir(target)
		if e != nil {
			return misc.Wrap(e, "read dir")
		}

		if len(entries) > 0 {
			logger.Warn().Logf("skip: directory not empty %s", target)
			logger.Logf("try: rm -fr %s", target)

			return nil
		}
	}

	if dryrun {
		_, err = fmt.Printf("remove %s\n", target)
		if err != nil {
			logger.Error().Logf("println: %v", err)
		}

		return nil
	}

	err = os.Remove(target)
	if err != nil {
		return misc.Wrap(err, "remove")
	}

	return nil
}

func bak(target string, dryrun bool) (err error) {
	bakFile := target + ".bak"

	_, err = os.Stat(bakFile)
	if err == nil {
		logger.Warn().Logf("skip: file exists %s", bakFile)

		return nil
	}

	fileStat, err := os.Stat(target)
	if err != nil {
		return misc.Wrap(err, "stat")
	}

	if fileStat.IsDir() {
		logger.Warn().Logf("skip: directory %s", target)
		logger.Logf("try: cp %s %s.bak", target, target)

		return nil
	}

	if dryrun {
		_, err = fmt.Printf("create %s.bak\n", target)
		if err != nil {
			logger.Error().Logf("println: %v", err)
		}

		return nil
	}

	backup, err := os.OpenFile(bakFile, os.O_CREATE|os.O_WRONLY, fileStat.Mode())
	if err != nil {
		return misc.Wrap(err, "open bak")
	}

	file, err := os.Open(target)
	if err != nil {
		return misc.Wrap(err, "open")
	}

	_, err = io.Copy(backup, file)
	if err != nil {
		return misc.Wrap(err, "copy")
	}

	return nil
}
