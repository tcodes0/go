package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var (
	//go:embed header.txt
	licenseHeader string
)

const (
	//nolint:varnamelen // go is the programming language
	Go int = iota + 1
	Shell
)

type Glob string

func (sourceFile Glob) String() string {
	return string(sourceFile)
}

func (sourceFile Glob) CommentToken() string {
	switch sourceFile.Kind() {
	default:
		return ""
	case Go:
		return "// "
	case Shell:
		return "# "
	}
}

func (sourceFile Glob) Kind() int {
	if strings.HasSuffix(sourceFile.String(), ".go") {
		return Go
	}

	if strings.HasSuffix(sourceFile.String(), ".sh") {
		return Shell
	}

	return 0
}

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fLogLevel := flagset.Int("log-level", int(logging.LInfo), "control logging output; 1 is debug, 5 is fatal; default 2 info.")
	fColor := flagset.Bool("color", false, "colored logging output; default false.")
	fGlobs := flagset.String("globs", "", "comma-space separated list of globs to search for files. Default empty.")
	fIgnore := flagset.String("ignore", "", "comma-space separated list of regexes to exclude files by path match. Default empty.")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("ERR: failed to parse flags: %v", err)
		os.Exit(1)
	}

	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(*fLogLevel))}
	if *fColor {
		opts = append(opts, logging.OptColor())
	}

	logger := logging.Create(opts...)
	globs := strings.Split(*fGlobs, ", ")
	rawRegexps := strings.Split(*fIgnore, ", ")

	if globs == nil {
		logger.Debug().Log("no globs provided")

		return
	}

	ignores := make([]*regexp.Regexp, 0, len(rawRegexps))

	for _, raw := range rawRegexps {
		var reg *regexp.Regexp

		if raw == "" {
			continue
		}

		reg, err = regexp.Compile(raw)
		if err != nil {
			logger.Fatalf("failed to compile regexp %s: %v", raw, err)
		}

		ignores = append(ignores, reg)
	}

	err = CopyrightBoilerplate(*logger, globs, ignores)
	if err != nil {
		logger.Fatalf("failed: %v", err)
	}
}

func CopyrightBoilerplate(logger logging.Logger, sourceFiles []string, ignoreRegexps []*regexp.Regexp) error {
	for _, sourceFile := range sourceFiles {
		matches, err := filepath.Glob(sourceFile)
		if err != nil {
			return misc.Wrap(err, "failed to glob files")
		}

		logger.Debug().Logf("files: %s, count %d", matches, len(matches))

		if len(matches) == 0 {
			logger.Log("no files matched")

			return nil
		}

		var headerWithComments string

	matchesLoop:
		for _, match := range matches {
			for _, regexp := range ignoreRegexps {
				if regexp.MatchString(match) {
					logger.Debug().Logf("skipping %s because ignore %s matches", match, regexp.String())

					continue matchesLoop
				}
			}

			if headerWithComments == "" {
				headerWithComments = addComments(licenseHeader, Glob(sourceFile))
			}

			hasHeader, content, err := checkForHeader(match, headerWithComments)
			if err != nil {
				return misc.Wrap(err, "failed to check for header")
			}

			if hasHeader {
				logger.Debug().Logf("header already present %s", match)

				break
			}

			err = applyHeader(headerWithComments, match, content, Glob(sourceFile))
			if err != nil {
				return misc.Wrap(err, "failed to apply header")
			}

			logger.Log(match)
		}
	}

	return nil
}

func checkForHeader(path, header string) (hasHeader bool, content string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return false, "", misc.Wrap(err, "failed to open file")
	}

	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return false, "", misc.Wrap(err, "failed to read file")
	}

	content = string(b)

	return strings.Contains(content, header), content, nil
}

func addComments(license string, glob Glob) (commentedLicense string) {
	for _, licenseLine := range strings.Split(license, "\n") {
		commentedLicense += glob.CommentToken() + licenseLine + "\n"
	}

	return commentedLicense
}

func applyHeader(header, path, content string, glob Glob) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return misc.Wrap(err, "unable to open file")
	}

	defer file.Close()

	switch glob.Kind() {
	default:
		return fmt.Errorf("unknown kind %d", glob.Kind())
	case Go:
		_, err = file.WriteString(header + "\n" + content)
	case Shell:
		shebang, rest, found := strings.Cut(content, "\n")
		if !found {
			_, err = file.WriteString(header + "\n" + shebang + "\n" + rest)
		} else {
			_, err = file.WriteString(shebang + "\n" + header + rest)
		}
	}

	if err != nil {
		return misc.Wrap(err, "unable to write to file")
	}

	return nil
}
