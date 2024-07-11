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
	"regexp"
	"slices"
	"strings"

	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"gopkg.in/yaml.v3"
)

const tMisc = "misc"

type config struct {
	Replace map[string]string `yaml:"replace"`
	Command string            `yaml:"command"`
	URL     string            `yaml:"url"`
}

var (
	//go:embed config.yml
	raw       []byte
	configs   config
	flagset   = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger    = &logging.Logger{}
	errUsage  = errors.New("see usage")
	RELogLine = regexp.MustCompile(`(?P<hash>[0-9a-f]+)\s+(?P<type>[a-zA-Z]+)(?:\((?P<scope>[^)]+)\))?:\s+(?P<description>.+)$`)
)

func main() {
	var err error

	// first deferred func will run last
	defer func() {
		if msg := recover(); msg != nil {
			logger.Stacktrace(true)
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
	_ = flagset.Bool("pizza", true, "pepperoni or mozzarella!. (default TRUE)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	err = yaml.Unmarshal(raw, &configs)
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	err = changelog()
}

func usage(err error) {
	if !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
	}

	fmt.Println("generate a markdown changelog from git log")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())
}

func changelog() error {
	// logLines, err := exec.Command(configs.Command).CombinedOutput()
	// if err != nil {
	// 	return misc.Wrapfl(err)
	// }
	logLines := `e0f4ee2 refactor(logging): threadsafety (#37)
933a1b2 refactor(cmd): remove exec (#36)
8cffa22 refactor(cmd/copyright): replace hardcoded configs with flags (#35)
d16eb9c feat: filer cmd (#29)
64535f1 refactor: re-write scripts in go (#28)
594740e refactor(copyright): use go cmd instead of script (#27)
fcc9cbb feat(license): automate copyright boilerplate header (#25)
50ae48b feat(ci): new checks & improvements to commitlint (#23)
dbed6c2 ci(commitlint): CI Commit lint job (#20)
4c40212 feat(modules): Add Modules (#13)
a2ddb28 feat(scripts): Tag script (#12)
332d5da docs: Update readme (#11)
5733201 docs(readme): Update readme (#6)
f3bbf73 feat(tooling): Improve tooling (#9)
af91878 feat(module): httpflush (#1)
992f6ec feat(ci): Github Workflows (#7)
3cc055c chore(repo): Reset main to remove test commits (#4)
3cc055c misc(repo): Reset main to remove test commits (#4)
`

	file, err := os.Open(".commitlintrc.yml")
	if err != nil {
		return misc.Wrapfl(err)
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return misc.Wrapfl(err)
	}

	commitlintrc := &struct {
		Rules map[string][]any `yaml:"rules"`
	}{}

	err = yaml.Unmarshal(fileContent, &commitlintrc)
	if err != nil {
		return misc.Wrapfl(err)
	}

	types, _ := (commitlintrc.Rules["type-enum"][2]).([]any)
	split := strings.Split(logLines, "\n")
	builder := strings.Builder{}
	typeBuilder := strings.Builder{}
	otherBuilder := strings.Builder{}
	scopeless := make([]string, 0, len(split))
	scoped := make([]string, 0, len(split))

	for _, t := range types {
		typ, _ := t.(string)

		for _, logLine := range split {
			match := RELogLine.FindStringSubmatch(logLine)
			if match == nil {
				if logLine != "" {
					logger.Warnf("log line does not match: %s", logLine)
				}

				continue
			}

			commitHash, lineType, scope, description := match[1], match[2], match[3], match[4]
			if lineType != typ {
				continue
			}

			if scope != "" {
				scoped = append(scoped, md("li", md("b", scope)+": "+description)+fmt.Sprintf(" (%s)\n", commitLink(commitHash)))
			} else {
				scopeless = append(scopeless, md("li", description)+fmt.Sprintf(" (%s)\n", commitLink(commitHash)))
			}
		}

		isOther := typ == tMisc || typ == "chore"

		if len(scoped) != 0 {
			slices.Sort(scoped)

			for _, s := range scoped {
				if isOther {
					otherBuilder.WriteString(s)
				} else {
					typeBuilder.WriteString(s)
				}
			}

			clear(scoped)
		}

		if len(scopeless) != 0 {
			for _, s := range scopeless {
				if isOther {
					otherBuilder.WriteString(s)
				} else {
					typeBuilder.WriteString(s)
				}
			}

			clear(scopeless)
		}

		if typeBuilder.Len() != 0 {
			prettyType, ok := configs.Replace[typ]
			if !ok {
				prettyType = typ
			}

			builder.WriteString(md("h2", prettyType) + "\n")
			builder.WriteString(typeBuilder.String())
			builder.WriteString("\n")
			typeBuilder.Reset()
		}
	}

	if otherBuilder.Len() != 0 {
		prettyType, ok := configs.Replace[tMisc]
		if !ok {
			prettyType = tMisc
		}

		builder.WriteString(md("h3", prettyType) + "\n")
		builder.WriteString(otherBuilder.String())
		builder.WriteString("\n")
		otherBuilder.Reset()
	}

	fmt.Print(builder.String())

	return nil
}

func md(tag, text string) string {
	switch tag {
	case "h1":
		return "# " + text
	case "h2":
		return "## " + text
	case "h3":
		return "### " + text
	case "h4":
		return "#### " + text
	case "li":
		return "- " + text
	case "b":
		return "**" + text + "**"
	}

	return text
}

func commitLink(hash string) string {
	return fmt.Sprintf("[%s](%s)", hash, configs.URL+"/commit/"+hash)
}
