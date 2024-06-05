package main

import (
	_ "embed"
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
	mockRegexp    = regexp.MustCompile(`/?mock_[^.]+\.go`)
)

const (
	//nolint:varnamelen // go is the programming language
	Go    int = iota + 1
	Shell int = iota + 1
)

type SourceFile struct {
	Glob         string
	CommentToken string
	Kind         int
}

func main() {
	logger := logging.Create(logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.LDebug), logging.OptColor())
	sourceFiles := []*SourceFile{
		{Glob: "./**/*.go", Kind: Go, CommentToken: "// "},
		{Glob: "./*.go", Kind: Go, CommentToken: "// "},
		{Glob: "./**/*.sh", Kind: Shell, CommentToken: "# "},
	}
	ignore := []*regexp.Regexp{mockRegexp}

	// TODO: parse from args
	err := CopyrightBoilerplate(*logger, sourceFiles, ignore)
	if err != nil {
		logger.Fatalf("failed: %v", err)
	}
}

func CopyrightBoilerplate(logger logging.Logger, sourceFiles []*SourceFile, ignoreRegexps []*regexp.Regexp) error {
	for _, sourceFile := range sourceFiles {
		matches, err := filepath.Glob(sourceFile.Glob)
		if err != nil {
			return misc.Wrap(err, "failed to glob files")
		}

		logger.Debug().Logf("files: %s, count %d", matches, len(matches))

		var headerWithComments string

		for _, match := range matches {
			for _, regexp := range ignoreRegexps {
				if regexp.MatchString(match) {
					logger.Debug().Logf("skipping %s because ignore %s matches", match, regexp.String())

					break
				}

				if headerWithComments == "" {
					headerWithComments = addComments(licenseHeader, sourceFile)
				}

				hasHeader, content, err := checkForHeader(match, headerWithComments)
				if err != nil {
					return misc.Wrap(err, "failed to check for header")
				}

				if hasHeader {
					logger.Debug().Logf("header already present %s", match)

					break
				}

				err = applyHeader(headerWithComments, match, content, sourceFile)
				if err != nil {
					return misc.Wrap(err, "failed to apply header")
				}

				logger.Log(match)
			}
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

func addComments(license string, file *SourceFile) (commentedLicense string) {
	for _, licenseLine := range strings.Split(license, "\n") {
		commentedLicense += file.CommentToken + licenseLine + "\n"
	}

	return commentedLicense
}

func applyHeader(header, path, content string, sourcefile *SourceFile) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		return misc.Wrap(err, "unable to open file")
	}

	defer file.Close()

	switch sourcefile.Kind {
	default:
		return fmt.Errorf("unknown kind %d", sourcefile.Kind)
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
