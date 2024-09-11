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
	"gopkg.in/yaml.v3"
)

const (
	actionLink   = "link"
	actionRemove = "remove"
	actionBackup = "backup"
)

type config struct {
	Version string `yaml:"version"`
}

var (
	//go:embed config.yml
	raw          string
	logger       = &logging.Logger{}
	flagset      = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	actions      = []string{actionLink, actionRemove, actionBackup}
	errUsage     = errors.New("see usage")
	errFinal     error
	descriptions = []string{
		"create symbolic link; provide files in pairs, one per line, first is source, second is link",
		"delete; one file per line",
		"copy with .bak extension; one file per line",
	}
)

func main() {
	defer func() {
		if msg := recover(); msg != nil {
			logger.Stacktrace(logging.LError, true)
			logger.Fatalf("%v", msg)
		}

		passAway(errFinal)
	}()

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	//nolint:gosec // level does not overflow here.
	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	fConfig := flagset.String("config", "", "path to config file (required)")
	fCommitL := flagset.Bool("commit", false, "apply changes (default: false)")
	fCommitS := flagset.Bool("c", false, "apply changes (default: false)")
	fVerShort := flagset.Bool("v", false, "print version and exit")
	fVerLong := flagset.Bool("version", false, "print version and exit")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		errFinal = err

		return
	}

	cfg := config{}

	err = yaml.Unmarshal([]byte(raw), &cfg)
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	if *fVerShort || *fVerLong {
		fmt.Println(cfg.Version)

		return
	}

	action, files, err := readConfig(*fConfig)
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	fDryrun := !(*fCommitL || *fCommitS)
	files = envVarResolver(files)
	errFinal = filer(files, action, fDryrun)
}

func passAway(fatal error) {
	if fatal != nil {
		if errors.Is(fatal, errUsage) || errors.Is(fatal, flag.ErrHelp) {
			usage(fatal)
		}

		logger.Stacktrace(logging.LDebug, true)
		logger.Fatalf("%s", fatal.Error())
	}
}

func usage(err error) {
	if !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
	}

	actionLines := []string{}

	for i, action := range actions {
		actionLines = append(actionLines, fmt.Sprintf("- %s: %s", action, descriptions[i]))
	}

	fmt.Printf(`Perform an action on a list of files.
First line of config file should be the action.
Changes nothing by default, pass -commit
Comments with # and newlines are ignored in config file

Actions available:
%s

%s
`, strings.Join(actionLines, "\n"), cmd.EnvVarUsage())
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
	var (
		count int
		errs  []error
	)

	// filerFns should do nothing if called with the wrong action.
	// Return patterns:
	// 0, nil: no changes.
	// len []error > 0: errors processing.
	// int > 0: changes made.
	type filerFn func(string, []string, bool) (int, []error)

	for _, fn := range []filerFn{link, remove, backup} {
		count, errs = fn(action, files, dryrun)
		if len(errs) > 0 {
			for _, err := range errs {
				logger.Errorf("%s", err.Error())
			}

			return errs[0]
		}

		if count > 0 {
			break
		}
	}

	if count == 0 {
		fmt.Printf("files ok")

		return nil
	}

	if dryrun {
		fmt.Println()
		fmt.Printf("use '%s -commit' to commit %d", strings.Join(os.Args, " "), count)
	} else {
		fmt.Printf("committed %d", count)
	}

	return nil
}

func link(action string, files []string, dryrun bool) (count int, errs []error) {
	if action != actionLink {
		return 0, nil
	}

	if len(files)%2 != 0 {
		return 0, []error{fmt.Errorf("link: file count not even: %d", len(files))}
	}

	for i, file := range files {
		if i%2 != 0 {
			continue
		}

		c, err := linkOne(file, files[i+1], dryrun)
		if err != nil {
			errs = append(errs, err)
		}

		count += c
	}

	return count, errs
}

func linkOne(source, link string, dryrun bool) (count int, err error) {
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

func remove(action string, files []string, dryrun bool) (count int, errs []error) {
	if action != actionRemove {
		return 0, nil
	}

	for _, file := range files {
		c, err := removeOne(file, dryrun)
		if err != nil {
			errs = append(errs, err)
		}

		count += c
	}

	return count, errs
}

func removeOne(target string, dryrun bool) (count int, err error) {
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

func backup(action string, files []string, dryrun bool) (count int, errs []error) {
	if action != actionBackup {
		return 0, nil
	}

	for _, file := range files {
		c, err := backupOne(file, dryrun)
		if err != nil {
			errs = append(errs, err)
		}

		count += c
	}

	return count, errs
}

func backupOne(target string, dryrun bool) (count int, err error) {
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
