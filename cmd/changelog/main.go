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
	"os/exec"
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
	URL     string            `yaml:"url"`
}

var (
	//go:embed config.yml
	raw      []byte
	configs  config
	flagset  = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger   = &logging.Logger{}
	errUsage = errors.New("see usage")
	//nolint:lll // long regex
	RELogLine = regexp.MustCompile(`(?P<hash>[0-9a-f]+)\s(?P<paren>\([^)]*\))?\s?(?P<type>[a-zA-Z]+)(?:\((?P<scope>[^)]+)\))?:\s(?P<description>.+)$`)
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
	logLines, types, err := prepare()
	if err != nil {
		return misc.Wrapfl(err)
	}

	builder, otherBuilder := buildChanges(types, logLines)
	if otherBuilder.Len() != 0 {
		prettyType, ok := configs.Replace[tMisc]
		if !ok {
			prettyType = tMisc
		}

		builder.WriteString(md("h4", prettyType) + "\n")
		builder.WriteString(otherBuilder.String())
		builder.WriteString("\n")
		otherBuilder.Reset()
	}

	fmt.Print(builder.String())

	return nil
}

func prepare() (logLines []string, types []any, err error) {
	byteLogLines, err := exec.Command("git", "log", "--oneline", "--decorate").Output()
	if err != nil {
		return nil, nil, misc.Wrapfl(err)
	}

	file, err := os.Open(".commitlintrc.yml")
	if err != nil {
		return nil, nil, misc.Wrapfl(err)
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, misc.Wrapfl(err)
	}

	commitlintrc := &struct {
		Rules map[string][]any `yaml:"rules"`
	}{}

	err = yaml.Unmarshal(fileContent, &commitlintrc)
	if err != nil {
		return nil, nil, misc.Wrapfl(err)
	}

	types, _ = (commitlintrc.Rules["type-enum"][2]).([]any)
	logLines = strings.Split(string(byteLogLines), "\n")

	for i, line := range logLines {
		if match := RELogLine.FindStringSubmatch(line); match != nil {
			if match[2] != "" && strings.Contains(match[2], "main") {
				// return lines for current branch only, stop at main
				return logLines[:i], types, nil
			}
		}
	}

	return logLines, types, nil
}

func buildChanges(types []any, logLines []string) (builder, otherBuilder *strings.Builder) {
	typeBuilder := &strings.Builder{}
	builder = &strings.Builder{}
	otherBuilder = &strings.Builder{}

	for _, t := range types {
		typ, _ := t.(string)
		scoped, scopeless := parseLine(logLines, typ)
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

	return builder, otherBuilder
}

func parseLine(lines []string, typ string) (scoped, scopeless []string) {
	scopeless = make([]string, 0, len(lines))
	scoped = make([]string, 0, len(lines))

	for _, line := range lines {
		match := RELogLine.FindStringSubmatch(line)
		if match == nil {
			if line != "" {
				logger.Warnf("no match: %s", line)
			}

			continue
		}

		commitHash, _, lineType, scope, description := match[1], match[2], match[3], match[4], match[5]
		if lineType != typ {
			continue
		}

		if scope != "" {
			scoped = append(scoped, md("li", md("b", scope)+": "+description)+fmt.Sprintf(" (%s)\n", commitLink(commitHash)))
		} else {
			scopeless = append(scopeless, md("li", description)+fmt.Sprintf(" (%s)\n", commitLink(commitHash)))
		}
	}

	return scoped, scopeless
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
