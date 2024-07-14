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
	"strconv"
	"strings"
	"time"

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

type changelogLine struct {
	Text     string
	Breaking bool
}

var (
	//go:embed config.yml
	raw       []byte
	configs   config
	flagset   = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger    = &logging.Logger{}
	errUsage  = errors.New("see usage")
	semverLen = 3
	//nolint:lll // long regex
	RELogLine = regexp.MustCompile(`(?P<hash>[0-9a-f]+)\s(?P<paren>\([^)]*\))?\s?(?P<type>[a-zA-Z]+)(?P<breaking1>!)?(?:\((?P<scope>[^)]+)\))?(?P<breaking2>!)?:\s(?P<description>.+)$`)
)

type semver []uint8

func (sv semver) String() string {
	if len(sv) < semverLen {
		return ""
	}

	return fmt.Sprintf("%d.%d.%d", sv[0], sv[1], sv[2])
}

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
	cfg := flagset.String("config", ".commitlintrc.yml", "path to commitlint config file")
	URL := flagset.String("url", "https://github.com/tcodes0/go", "github repository URL to generate commit links")
	module := flagset.String("module", "", "module changed, used for changelog title (required)")

	err = flagset.Parse(os.Args[1:])
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	if *module == "" {
		err = errors.Join(errors.New("module required"), errUsage)

		return
	}

	err = yaml.Unmarshal(raw, &configs)
	if err != nil {
		err = errors.Join(err, errUsage)

		return
	}

	err = changelog(*cfg, *URL, *module)
}

func usage(err error) {
	if !errors.Is(err, flag.ErrHelp) {
		flagset.Usage()
	}

	fmt.Println("generate a markdown changelog from git log")
	fmt.Println("a prior tag with format module/vX.Y.Z must exist on main")
	fmt.Println("unstable tags (0.x.x) will not be promoted to 1.0.0 automaticall, do it manually")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())
}

func changelog(cfg, url, module string) error {
	byteLogLines, err := exec.Command("git", "log", "--oneline", "--decorate").Output()
	if err != nil {
		return misc.Wrapfl(err)
	}

	logLines := strings.Split(string(byteLogLines), "\n")

	logLines, oldVer, newVer, err := parseGitLog(module, logLines)
	if err != nil {
		return misc.Wrapfl(err)
	}

	if oldVer == "" || newVer == "" {
		return misc.Wrapfl(fmt.Errorf("malformed version: old: %s, new: %s", oldVer, newVer))
	}

	types, err := parseConfig(cfg)
	if err != nil {
		return misc.Wrapfl(err)
	}

	builder := &strings.Builder{}
	title := fmt.Sprintf("%s: v%s %s\n\n", module, newVer, md("i", "("+time.Now().Format("2006-01-02")+")"))
	builder.WriteString(md("h1", title))

	if newVer != "" && oldVer != "" {
		builder.WriteString(md("h3", compareLink(url, newVer, oldVer)) + "\n\n")
	}

	miscBuilder := changes(types, logLines, url, builder)
	if miscBuilder.Len() != 0 {
		prettyType, ok := configs.Replace[tMisc]
		if !ok {
			prettyType = tMisc
		}

		builder.WriteString(md("h4", prettyType) + "\n")
		builder.WriteString(miscBuilder.String())
		builder.WriteString("\n")
		miscBuilder.Reset()
	}

	fmt.Print(builder.String())

	return nil
}

func parseGitLog(module string, allLogLines []string) (branchLogLines []string, versionOld, versionNew string, err error) {
	oldVer := make(semver, 0, semverLen)
	REReleaseTag := regexp.MustCompile("tag: " + module + `\/v(?P<version>\d+\.\d+\.\d+)`)

	for _, line := range allLogLines {
		if match := REReleaseTag.FindStringSubmatch(line); match != nil {
			for _, versionN := range strings.Split(match[1], ".") {
				version, err := strconv.ParseInt(versionN, 10, 8)
				if err != nil {
					return nil, "", "", misc.Wrapfl(err)
				}

				oldVer = append(oldVer, uint8(version))
			}
		}
	}

	if len(oldVer) == 0 {
		return nil, "", "", misc.Wrapfl(fmt.Errorf("tag not found: %s/vx.x.x", module))
	}

	REMinor := regexp.MustCompile(`feat:|feat\(.+\):`)
	branchLogLines = make([]string, 0, len(allLogLines))
	breaking, minor := false, false

	for i, line := range allLogLines {
		if match := RELogLine.FindStringSubmatch(line); match != nil {
			if len(branchLogLines) == 0 && match[2] != "" && strings.Contains(match[2], "main") {
				// seeing "main" means the current branch log ended
				branchLogLines = allLogLines[:i]

				break
			}

			if !breaking {
				breaking = match[4] != "" || match[6] != ""
			}

			if !minor && !breaking {
				minor = REMinor.MatchString(line)
			}
		}
	}

	newVer := versionUp(oldVer, oldVer[0] == 0, breaking, minor)

	return branchLogLines, oldVer.String(), newVer.String(), nil
}

func versionUp(current semver, unstable, breaking, minor bool) semver {
	newVer := make(semver, semverLen)
	copy(newVer, current)

	if unstable {
		if breaking {
			newVer[1]++
			newVer[2] = 0
		} else {
			newVer[2]++
		}

		return newVer
	}

	if breaking {
		newVer[0]++
		newVer[1] = 0
		newVer[2] = 0
	} else if minor {
		newVer[1]++
		newVer[2] = 0
	} else {
		newVer[2]++
	}

	return newVer
}

func parseConfig(cfg string) (types []any, err error) {
	file, err := os.Open(cfg)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	commitlintrc := &struct {
		Rules map[string][]any `yaml:"rules"`
	}{}

	err = yaml.Unmarshal(fileContent, &commitlintrc)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	types, _ = (commitlintrc.Rules["type-enum"][2]).([]any)

	if len(types) == 0 {
		return nil, fmt.Errorf("empty %s.type-enum[2]", cfg)
	}

	return types, nil
}

func changes(types []any, logLines []string, url string, builder *strings.Builder) *strings.Builder {
	typeBuilder := &strings.Builder{}
	miscBuilder := &strings.Builder{}

	for _, t := range types {
		typ, _ := t.(string)
		scoped, scopeless, breakings := parseLines(logLines, typ, url)

		if len(scoped) != 0 {
			slices.SortFunc(scoped, sortFn)
			writeLines(scoped, typ, typeBuilder, miscBuilder)
		}

		if len(scopeless) != 0 {
			writeLines(scopeless, typ, typeBuilder, miscBuilder)
		}

		if len(breakings) != 0 {
			builder.WriteString(md("h2", "Breaking Changes") + "\n")

			for _, b := range breakings {
				builder.WriteString(b.Text)
			}

			builder.WriteString("\n")
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

	return miscBuilder
}

func sortFn(a, b changelogLine) int {
	if a.Text < b.Text {
		return -1
	} else if a.Text > b.Text {
		return 1
	}

	return 0
}

func writeLines(lines []changelogLine, typ string, typeBuilder, otherBuilder *strings.Builder) {
	for _, s := range lines {
		if typ == tMisc || typ == "chore" {
			otherBuilder.WriteString(s.Text)
		} else {
			typeBuilder.WriteString(s.Text)
		}
	}

	clear(lines)
}

func parseLines(lines []string, typ, url string) (scoped, scopeless, breakings []changelogLine) {
	scopeless = make([]changelogLine, 0, len(lines))
	scoped = make([]changelogLine, 0, len(lines))
	breakings = make([]changelogLine, 0, len(lines))

	for _, line := range lines {
		match := RELogLine.FindStringSubmatch(line)
		if match == nil {
			if line != "" {
				logger.Errorf("skip, no match: %s", line)
			}

			continue
		}

		commitHash, _, lineType, breaking1, scope, breaking2, description := match[1], match[2], match[3], match[4], match[5], match[6], match[7]
		if lineType != typ {
			continue
		}

		cline := changelogLine{}

		if breaking1 != "" || breaking2 != "" {
			cline.Breaking = true
		}

		if scope != "" {
			cline.Text = md("li", md("b", scope)+": "+description) + fmt.Sprintf(" (%s)\n", commitLink(url, commitHash))
			if !cline.Breaking {
				scoped = append(scoped, cline)
			}
		} else {
			cline.Text = md("li", description) + fmt.Sprintf(" (%s)\n", commitLink(url, commitHash))
			if !cline.Breaking {
				scopeless = append(scopeless, cline)
			}
		}

		if cline.Breaking {
			breakings = append(breakings, cline)
		}
	}

	return scoped, scopeless, breakings
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
	case "i":
		return "*" + text + "*"
	}

	return text
}

func commitLink(url, hash string) string {
	return fmt.Sprintf("[%s](%s/commit/%s)", hash, url, hash)
}

func compareLink(url, tag1, tag2 string) string {
	return fmt.Sprintf("[Diff with %s](%s/compare/%s..%s)", tag2, url, tag2, tag1)
}
