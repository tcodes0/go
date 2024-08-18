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
	actionBackup = "backup"
)

var (
	logger       = &logging.Logger{}
	flagset      = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	actions      = []string{actionLink, actionRemove, actionBackup}
	errUsage     = errors.New("see usage")
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
			if errors.Is(err, errUsage) {
				usage(err)
			}

			logger.Fatalf("%s", err.Error())
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
	fConfig := flagset.String("config", "", "path to config file (required)")
	fCommitL := flagset.Bool("commit", false, "apply changes (default: false)")
	fCommitS := flagset.Bool("c", false, "apply changes (default: false)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	fDryrun := !(*fCommitL || *fCommitS)

	action, files, err := readConfig(*fConfig)
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	files = envVarResolver(files)
	err = filer(files, action, fDryrun)
}

func readConfig(configPath string) (action string, lines []string, err error) {
	if configPath == "" {
		return "", nil, errors.New("empty config")
	}

	cfgFile, err := os.Open(configPath)
	if err != nil {
		return "", nil, misc.Wrapfl(err)
	}

	raw, err := io.ReadAll(cfgFile)
	if err != nil {
		return "", nil, misc.Wrapfl(err)
	}

	lines = strings.Split(string(raw), "\n")
	lines = lo.Filter(lines, func(file string, _ int) bool {
		return file != "" && !strings.HasPrefix(file, "#")
	})

	if len(lines) == 0 {
		return "", nil, errors.New("empty config")
	}

	action, files := lines[0], lines[1:]

	if _, found := lo.Find(actions, func(a string) bool { return a == action }); !found {
		return "", nil, misc.Wrapf(errUsage, "unknown action %s", action)
	}

	return action, files, nil
}

func usage(err error) {
	if !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
	}

	fmt.Println("perform an action on a list of files")
	fmt.Println("first line of config file should be the action")
	fmt.Println("changes nothing by default, pass -commit")
	fmt.Println("comments with # and newlines are ignored in config file")
	fmt.Println()
	fmt.Println("actions available:")

	for i, action := range actions {
		fmt.Printf("- %s: %s\n", action, descriptions[i])
	}

	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())
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
	errs, count := []error{}, 0

	switch action {
	case actionLink:
		if len(files)%2 != 0 {
			return fmt.Errorf("link: file count not even: %d", len(files))
		}

		for i, file := range files {
			if i%2 != 0 {
				continue
			}

			c, err := link(file, files[i+1], dryrun)
			if err != nil {
				errs = append(errs, err)
			}

			count += c
		}
	case actionRemove:
		for _, file := range files {
			c, err := remove(file, dryrun)
			if err != nil {
				errs = append(errs, err)
			}

			count += c
		}
	case actionBackup:
		for _, file := range files {
			c, err := backup(file, dryrun)
			if err != nil {
				errs = append(errs, err)
			}

			count += c
		}
	}

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Errorf("%s", err.Error())
		}

		return errs[0]
	}

	if count == 0 {
		fmt.Println("files ok")

		return nil
	}

	if dryrun {
		fmt.Printf("to commit %d changes run: %s -commit", count, strings.Join(os.Args, " "))
	}

	return nil
}

func link(source, link string, dryrun bool) (count int, err error) {
	_, err = os.Stat(source)
	if err != nil {
		return 0, misc.Wrapf(err, "stat")
	}

	linkDir := filepath.Dir(link)

	_, err = os.Stat(linkDir)
	if err != nil {
		return 0, misc.Wrapf(err, "try: mkdir -p %s", linkDir)
	}

	// do not follow symlinks!
	lStat, err := os.Lstat(link)
	if err == nil {
		logger.Debugf("skip: file exists %s", link)

		if lStat.Mode()&os.ModeSymlink != 0 {
			_, err = os.Stat(link)
			if err != nil {
				logger.Errorf("broken link; try: unlink %s", link)
			}
		}

		return 0, nil
	}

	if dryrun {
		_, err = fmt.Printf("link %s -> %s\n", link, source)
		if err != nil {
			logger.Errorf("println: %v", err)
		}

		return 1, nil
	}

	err = os.Symlink(source, link)
	if err != nil {
		return 0, misc.Wrapf(err, "symlink")
	}

	return 1, nil
}

func remove(target string, dryrun bool) (count int, err error) {
	stat, err := os.Stat(target)
	if err != nil {
		logger.Debugf("skip: file not found %s", target)

		//nolint:nilerr // func about removing files
		return 0, nil
	}

	if stat.IsDir() {
		entries, e := os.ReadDir(target)
		if e != nil {
			return 0, misc.Wrap(e, "read dir")
		}

		if len(entries) > 0 {
			return 0, fmt.Errorf("directory not empty; try: rm -fr %s", target)
		}
	}

	if dryrun {
		_, err = fmt.Printf("remove %s\n", target)
		if err != nil {
			logger.Errorf("println: %v", err)
		}

		return 1, nil
	}

	err = os.Remove(target)
	if err != nil {
		return 0, misc.Wrap(err, "remove")
	}

	return 1, nil
}

func backup(target string, dryrun bool) (count int, err error) {
	bakFile := target + ".bak"

	_, err = os.Stat(bakFile)
	if err == nil {
		logger.Debugf("skip: file exists %s", bakFile)

		return 0, nil
	}

	fileStat, err := os.Stat(target)
	if err != nil {
		return 0, misc.Wrap(err, "stat")
	}

	if fileStat.IsDir() {
		return 0, fmt.Errorf("not a file; try: cp %s %s.bak", target, target)
	}

	if dryrun {
		_, err = fmt.Printf("create %s.bak\n", target)
		if err != nil {
			logger.Errorf("println: %v", err)
		}

		return 1, nil
	}

	backup, err := os.OpenFile(bakFile, os.O_CREATE|os.O_WRONLY, fileStat.Mode())
	if err != nil {
		return 0, misc.Wrap(err, "open bak")
	}

	file, err := os.Open(target)
	if err != nil {
		return 0, misc.Wrap(err, "open")
	}

	_, err = io.Copy(backup, file)
	if err != nil {
		return 0, misc.Wrap(err, "copy")
	}

	return 1, nil
}
