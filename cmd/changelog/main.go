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

	"github.com/samber/lo"
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
	Hash     string
	Breaking bool
	Minor    bool
}

var (
	//go:embed config.yml
	raw       []byte
	configs   config
	flagset   = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	logger    = &logging.Logger{}
	errUsage  = errors.New("see usage")
	semverLen = 3
	//nolint:lll // regex
	RECommitLine = regexp.MustCompile(`^(?P<asterisk>\*? ?)(?P<type>[a-zA-Z]+)(?P<breaking1>!)?(?:\((?P<scope>[^)]+)\))?(?P<breaking2>!)?:\s(?P<description>.+?)$`)
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
	fmt.Println("a prior tag with format module/vx.x.x must exist on main")
	fmt.Println("unstable tags (0.x.x) will not be promoted to 1.0.0 automatically, do it manually")
	fmt.Println()
	fmt.Println(cmd.EnvVarUsage())
}

func changelog(cfg, url, module string) error {
	mods, err := cmd.FindModules(logger)
	if err != nil {
		return misc.Wrapfl(err)
	}

	_, found := lo.Find(mods, func(m string) bool { return m == module })
	if !found {
		return fmt.Errorf("unknown module: %s", module)
	}

	// format: hash\n (tags branches)\ncommit message and body in multiple lines\n
	byteLogLines, err := exec.Command("git", "log", "--pretty=format:%H%n%d%n%B").Output()
	if err != nil {
		return misc.Wrapfl(err)
	}

	splitLines := strings.Split(string(byteLogLines), "\n")

	releaseLines, oldVer, err := parseGitLog(module, splitLines)
	if err != nil {
		return misc.Wrapfl(err)
	}

	types, err := parseConfig(cfg)
	if err != nil {
		return misc.Wrapfl(err)
	}

	document := &strings.Builder{}

	newVer, body, footer := writeContent(types, releaseLines, oldVer, url)
	title := fmt.Sprintf("%s: v%s %s\n\n", module, newVer, md("i", "("+time.Now().Format("2006-01-02")+")"))
	document.WriteString(md("h1", title))
	document.WriteString(md("h3", compareLink(url, tag(module, newVer.String()), tag(module, oldVer.String()))) + "\n\n")

	if body.Len() != 0 {
		document.WriteString(body.String())
		body.Reset()
	}

	if footer.Len() != 0 {
		prettyType, ok := configs.Replace[tMisc]
		if !ok {
			prettyType = tMisc
		}

		document.WriteString(md("h4", prettyType) + "\n")
		document.WriteString(footer.String())
		document.WriteString("\n")
		footer.Reset()
	}

	fmt.Print(document.String())

	return nil
}

func parseGitLog(module string, allLogLines []string) (releaseLogLines []changelogLine, versionOld semver, err error) {
	oldVer := make(semver, 0, semverLen)
	releaseLogLines = make([]changelogLine, 0, len(allLogLines))
	REReleaseTag := regexp.MustCompile("tag: " + module + `\/v(?P<version>\d+\.\d+\.\d+)`)
	REHash := regexp.MustCompile(`^[abcdef0-9]+$`)
	hash := ""

	for _, line := range allLogLines {
		line = strings.TrimSpace(line)
		logger.Debugf("line=%s", line)

		if line == "" {
			continue
		}

		if match := REHash.FindString(line); match != "" {
			// save the hash for the commit lines that follow
			hash = match

			continue
		}

		if match := RECommitLine.FindStringSubmatch(line); match != nil {
			if match[1] /*asterisk*/ == "" {
				// commit head has no asterisk, body lines do.
				// commit head is repeated in the body, skip.
				continue
			}

			releaseLogLines = append(releaseLogLines, changelogLine{Text: line, Hash: hash})
		}

		if match := REReleaseTag.FindStringSubmatch(line); match != nil {
			for _, versionN := range strings.Split(match[1], ".") {
				version, err := strconv.ParseInt(versionN, 10, 8)
				if err != nil {
					return nil, nil, misc.Wrapfl(err)
				}

				oldVer = append(oldVer, uint8(version))
			}

			// seeing a module tag means it is the old tag,
			// the release log ended and we are done
			break
		}
	}

	if len(oldVer) == 0 {
		return nil, nil, fmt.Errorf("tag not found: %s", tag(module, "?.?.?"))
	}

	return releaseLogLines, oldVer, nil
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

func writeContent(types []any, logLines []changelogLine, oldVer semver, url string) (newVer semver, body, footer *strings.Builder) {
	footer, body = &strings.Builder{}, &strings.Builder{}
	minor, breaks := false, false

	for _, t := range types {
		var scoped, scopeless, breakings []changelogLine

		typ, _ := t.(string)
		paragraph := &strings.Builder{}
		scoped, scopeless, breakings, minor = parseLines(logLines, typ, url)

		if len(scoped) != 0 {
			slices.SortFunc(scoped, sortFn)
			writeLines(scoped, typ, paragraph, footer)
		}

		if len(scopeless) != 0 {
			writeLines(scopeless, typ, paragraph, footer)
		}

		if len(breakings) != 0 {
			breaks = true

			body.WriteString(md("h2", "Breaking Changes") + "\n")

			for _, b := range breakings {
				body.WriteString(b.Text)
			}

			body.WriteString("\n")
		}

		if paragraph.Len() != 0 {
			prettyType, ok := configs.Replace[typ]
			if !ok {
				prettyType = typ
			}

			body.WriteString(md("h2", prettyType) + "\n")
			body.WriteString(paragraph.String())
			body.WriteString("\n")
			paragraph.Reset()
		}
	}

	newVer = versionUp(oldVer, oldVer[0] == 0, breaks, minor)

	return newVer, body, footer
}

func parseLines(lines []changelogLine, typ, url string) (scoped, scopeless, breakings []changelogLine, minor bool) {
	scopeless = make([]changelogLine, 0, len(lines))
	scoped = make([]changelogLine, 0, len(lines))
	breakings = make([]changelogLine, 0, len(lines))

	for _, line := range lines {
		match := RECommitLine.FindStringSubmatch(line.Text)
		if match == nil {
			if line.Text != "" {
				logger.Errorf("skip, no match: %s", line.Text)
			}

			continue
		}

		_, lineType, breaking1, scope, breaking2, description := match[1], match[2], match[3], match[4], match[5], match[6]
		if lineType != typ {
			continue
		}

		line.Breaking = breaking1 != "" || breaking2 != ""
		line.Minor = lineType == "feat"

		if scope != "" {
			line.Text = md("li", md("b", scope)+": "+description) + fmt.Sprintf(" (%s)\n", commitLink(url, line.Hash))
			if !line.Breaking {
				scoped = append(scoped, line)
			}
		} else {
			line.Text = md("li", description) + fmt.Sprintf(" (%s)\n", commitLink(url, line.Hash))
			if !line.Breaking {
				scopeless = append(scopeless, line)
			}
		}

		if line.Breaking {
			breakings = append(breakings, line)
		}
	}

	return scoped, scopeless, breakings, minor
}

func sortFn(a, b changelogLine) int {
	if a.Text < b.Text {
		return -1
	} else if a.Text > b.Text {
		return 1
	}

	return 0
}

func writeLines(lines []changelogLine, typ string, paragraph, footer *strings.Builder) {
	for _, s := range lines {
		if typ == tMisc || typ == "chore" {
			footer.WriteString(s.Text)
		} else {
			paragraph.WriteString(s.Text)
		}
	}

	clear(lines)
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
	return fmt.Sprintf("[%s](%s/commit/%s)", hash[:8], url, hash)
}

func tag(module, version string) string {
	return module + "/v" + version
}

func compareLink(url, tag1, tag2 string) string {
	return fmt.Sprintf("[Diff with %s](%s/compare/%s..%s)", tag2, url, tag2, tag1)
}
