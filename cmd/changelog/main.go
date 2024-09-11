// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/tcodes0/go/cmd"
	"github.com/tcodes0/go/cmd/changelog/github"
	"github.com/tcodes0/go/jsonutil"
	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

const tMisc = "misc"

type config struct {
	Replace map[string]string `yaml:"replace"`
	URL     string            `yaml:"url"`
	Version string            `yaml:"version"`
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
	//nolint:lll // regex https://regex101.com/r/tYwQcJ/3
	RECommitLine = regexp.MustCompile(`^(?P<asterisk>\*? ?)(?P<type>[a-zA-Z]+)(?P<breaking1>!)?(?:\((?P<scope>[^)]+)\))?(?P<breaking2>!)?:\s(?P<description>.+?)(?:\s\(#(?P<PR>\d+)\))?$`)
	errFinal     error
)

type semver []uint8

func (sv semver) String() string {
	if len(sv) < semverLen {
		return ""
	}

	return fmt.Sprintf("%d.%d.%d", sv[0], sv[1], sv[2])
}

func main() {
	defer func() {
		passAway(errFinal)
	}()

	misc.DotEnv(".env", false /* noisy */)

	fColor := misc.LookupEnv(cmd.EnvColor, false)
	fLogLevel := misc.LookupEnv(cmd.EnvLogLevel, int(logging.LInfo))

	//nolint:gosec // log level fits in uint8.
	opts := []logging.CreateOptions{logging.OptFlags(log.Lshortfile), logging.OptLevel(logging.Level(fLogLevel))}
	if fColor {
		opts = append(opts, logging.OptColor())
	}

	logger = logging.Create(opts...)
	cfg := flagset.String("config", ".commitlintrc.yml", "path to commitlint config file")
	title := flagset.String("title", "", "release title; new version and date will be added")
	tagPrefixRaw := flagset.String("tagprefixes", "", "comma separated prefixes to find tags, i.e $PREFIXv1.0.0")
	repoURL := flagset.String("url", "", "github repository URL to point links at, prefixed 'https://github.com/' (required)")
	tagsFile := flagset.String("tagsfile", "", "write tags to file")
	fVerShort := flagset.Bool("v", false, "print version and exit")
	fVerLong := flagset.Bool("version", false, "print version and exit")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	err = yaml.Unmarshal(raw, &configs)
	if err != nil {
		errFinal = errors.Join(err, errUsage)

		return
	}

	if *fVerShort || *fVerLong {
		fmt.Println(configs.Version)

		return
	}

	if *repoURL == "" {
		errFinal = errors.Join(errors.New("url required"), errUsage)

		return
	}

	prefixes := strings.Split(strings.ReplaceAll(*tagPrefixRaw, ", ", ","), ",")

	errFinal = changelog(*cfg, *repoURL, *title, *tagsFile, prefixes)
}

// Defer from main() very early; the first deferred function will run last.
// Gracefully handles panics and fatal errors. Replaces os.exit(1).
func passAway(fatal error) {
	if msg := recover(); msg != nil {
		logger.Stacktrace(logging.LError, true)
		logger.Fatalf("%v", msg)
	}

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

	fmt.Printf(`generate a markdown changelog from git log
a prior tag with format ${PREFIX}vx.x.x must exist on main
unstable tags (0.x.x) will not be promoted to 1.0.0 automatically, do it manually

%s
`, cmd.EnvVarUsage())
}

func changelog(cfg, repoURL, title, tagsFile string, tagPrefixes []string) error {
	err := validateInputs(tagPrefixes)
	if err != nil {
		return misc.Wrapfl(err)
	}

	// format: hash\n (tags branches)\ncommit message and body in multiple lines\n
	byteLogLines, err := exec.Command("git", "log", "--pretty=format:%H%n%d%n%B").Output()
	if err != nil {
		return misc.Wrapfl(err)
	}

	splitLines := strings.Split(string(byteLogLines), "\n")

	releaseLines, oldVers, prs, err := parseGitLog(tagPrefixes, splitLines)
	if err != nil {
		return misc.Wrapfl(err)
	}

	if len(prs) != 0 {
		var prLines []changelogLine

		logger.DebugData(map[string]any{"prs": prs, "url": repoURL}, "querying github")

		prLines, err = fetchPullRequests(prs, repoURL)
		if err != nil {
			logger.Error(err)
			logger.Info("changelog will be generated without github information")
		} else {
			releaseLines = prLines
		}
	}

	types, err := parseConfig(cfg)
	if err != nil {
		return misc.Wrapfl(err)
	}

	titleColon := ": "
	if title == "" {
		titleColon = ""
	}

	doc, newVers := writeDocument(types, releaseLines, oldVers, prs, repoURL, title, titleColon, tagPrefixes)
	fmt.Print(doc)

	if tagsFile != "" {
		err = writeTags(tagsFile, tagPrefixes, newVers)
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func validateInputs(tagPrefixes []string) error {
	REPrefixRaw := `^[a-zA-Z0-9\/]+$`
	RETag := regexp.MustCompile(REPrefixRaw)

	for _, tagPrefix := range tagPrefixes {
		if tagPrefix != "" && !RETag.MatchString(tagPrefix) {
			return fmt.Errorf("invalid tag prefix: '%s', should match %s", tagPrefix, REPrefixRaw)
		}
	}

	return nil
}

func parseGitLog(tagPrefixes, allLogLines []string) (releaseLogLines []changelogLine, versionOld []semver, prs []string, err error) {
	releaseLogLines = make([]changelogLine, 0, len(allLogLines))
	REHash, oldVersions, hash := regexp.MustCompile(`^[abcdef0-9]+$`), make([]semver, 0, len(tagPrefixes)), ""

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
				// if line has asterisk it is a commit in the body, else it is a pull request title,
				// from the pull request title only the PR number is important.
				if match[7] /*PR*/ != "" {
					prs = append(prs, match[7])
				}

				continue
			}

			releaseLogLines = append(releaseLogLines, changelogLine{Text: line, Hash: hash})
		}

		if strings.Contains(line, "tag:") {
			parsedOldVers, err := parseTags(tagPrefixes, line)
			if err != nil {
				return nil, nil, nil, misc.Wrapfl(err)
			}

			oldVersions = slices.Concat(oldVersions, parsedOldVers)
		}

		if len(oldVersions) == len(tagPrefixes) {
			// found all the old tags
			break
		}
	}

	if len(oldVersions) != len(tagPrefixes) {
		return nil, nil, nil, fmt.Errorf("wanted old tags: %#v, found: %#v", tagPrefixes, oldVersions)
	}

	slices.Reverse(prs) // reads better in changelog

	return releaseLogLines, oldVersions, prs, nil
}

func parseTags(tagPrefixes []string, line string) (oldVersions []semver, err error) {
	for _, tagPrefix := range tagPrefixes {
		REReleaseTag := regexp.MustCompile("tag: " + tagPrefix + `v(?P<version>\d+\.\d+\.\d+)`)

		if match := REReleaseTag.FindStringSubmatch(line); match != nil {
			oldVer := make(semver, 0, semverLen)

			for _, versionN := range strings.Split(match[1], ".") {
				version, err := strconv.ParseInt(versionN, 10, 8)
				if err != nil {
					return nil, misc.Wrapfl(err)
				}

				//nolint:gosec // parse int validates to int8
				oldVer = append(oldVer, uint8(version))
			}

			oldVersions = append(oldVersions, oldVer)
		}
	}

	return oldVersions, nil
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

func fetchPullRequests(prs []string, ghURL string) (lines []changelogLine, err error) {
	userRepo, limit, query, header := strings.TrimPrefix(ghURL, "https://github.com/"), 100, url.Values{}, http.Header{}
	fatCommits, group, lock := map[int]*[]*github.FatCommit{}, errgroup.Group{}, sync.Mutex{}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("empty GITHUB_TOKEN env var")
	}

	query.Add("page", "1")
	query.Add("per_page", strconv.Itoa(limit))
	header.Add("Accept", "application/vnd.github.v3+json")
	header.Add("Authorization", "token "+token)

	for i, pullReq := range prs {
		group.Go(func() error {
			ctx := logger.WithContext(context.TODO())

			fats, e := fetchOne(ctx, fmt.Sprintf("repos/%s/pulls/%s/commits?%s", userRepo, pullReq, query.Encode()), header, nil)
			if e != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()

			fatCommits[i] = fats

			if len(*fats) == limit {
				return fmt.Errorf("PR #%s requires pagination", pullReq)
			}

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	for _, fats := range fatCommits {
		for _, fat := range *fats {
			lines = append(lines, changelogLine{Text: "* " + fat.Commit.Message, Hash: fat.SHA})
		}
	}

	return lines, nil
}

func fetchOne(ctx context.Context, path string, header http.Header, client *http.Client) (*[]*github.FatCommit, error) {
	res, err := github.Get(ctx, path, header, client)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", res.StatusCode)
	}

	fatCommits, err := jsonutil.UnmarshalReader[[]*github.FatCommit](res.Body)
	res.Body.Close()

	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	return fatCommits, nil
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

func writeDocument(types []any, releaseLines []changelogLine, oldVers []semver, prs []string, repoURL, title, titleColon string,
	tagPrefixes []string,
) (doc string, newVers []semver) {
	document := &strings.Builder{}
	newVers, header, body, footer := compose(types, releaseLines, oldVers, repoURL)

	for i, tagPrefix := range tagPrefixes {
		titleH1 := fmt.Sprintf("%s%s%sv%s %s\n", title, titleColon, tagPrefix, newVers[i], md("i", "("+time.Now().Format("2006-01-02")+")"))
		document.WriteString(md("h1", titleH1))
		document.WriteString(md("h3",
			link("Diff with "+tag(tagPrefix, oldVers[i].String()),
				fmt.Sprintf("%s/compare/%s..%s", repoURL, tag(tagPrefix, newVers[i].String()), tag(tagPrefix, oldVers[i].String()))),
		) + "\n")
	}

	document.WriteString("\n")

	prLinks := lo.Map(prs, func(pr string, _ int) string { return link("#"+pr, fmt.Sprintf("%s/pull/%s", repoURL, pr)) })
	document.WriteString(md("h3", "PRs in this release: "+strings.Join(prLinks, ", ")+"\n"))

	if header.Len() != 0 {
		document.WriteString(header.String() + "\n")
		header.Reset()
	}

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
		document.WriteString(footer.String() + "\n")
		footer.Reset()
	}

	return document.String(), newVers
}

func compose(types []any, logLines []changelogLine, oldVers []semver, repoURL string) (
	newVers []semver, body, footer, header *strings.Builder,
) {
	footer, body, header = &strings.Builder{}, &strings.Builder{}, &strings.Builder{}
	// limitation: minor and breaks will apply to all versions, consequence of releasing many tags together
	minor, breaks := false, false
	// removes repetitive commits like 'misc: fix ci' even if the hashes are different
	uniqLines := lo.UniqBy(logLines, func(line changelogLine) string { return line.Text })

	for _, t := range types {
		var scoped, scopeless, breakings []changelogLine

		typ, _ := t.(string)
		paragraph := &strings.Builder{}
		scoped, scopeless, breakings, minor = parseLines(uniqLines, typ, repoURL)

		if len(scoped) != 0 {
			slices.SortFunc(scoped, sortFn)
			writeLines(scoped, typ, paragraph, footer)
		}

		if len(scopeless) != 0 {
			writeLines(scopeless, typ, paragraph, footer)
		}

		if len(breakings) != 0 {
			if !breaks {
				header.WriteString(md("h2", md("i", "Breaking Changes")) + "\n")

				breaks = true
			}

			for _, b := range breakings {
				header.WriteString(b.Text)
			}
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

	newVers = make([]semver, len(oldVers))

	for i, oldVer := range oldVers {
		newVers[i] = versionUp(oldVer, oldVer[0] == 0, breaks, minor)
	}

	return newVers, header, body, footer
}

func parseLines(lines []changelogLine, typ, repoURL string) (scoped, scopeless, breakings []changelogLine, minor bool) {
	scopeless = make([]changelogLine, 0, len(lines))
	scoped = make([]changelogLine, 0, len(lines))
	breakings = make([]changelogLine, 0, len(lines))

	for _, line := range lines {
		match := RECommitLine.FindStringSubmatch(line.Text)
		if match == nil {
			if line.Text != "" {
				logger.Warnf("ignore, no match: %s", line.Text)
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
			line.Text = md("li", md("b", scope)+": "+description) + fmt.Sprintf(" (%s)\n",
				link(line.Hash[:8], fmt.Sprintf("%s/commit/%s", repoURL, line.Hash)),
			)
			if !line.Breaking {
				scoped = append(scoped, line)
			}
		} else {
			line.Text = md("li", description) + fmt.Sprintf(" (%s)\n", link(line.Hash[:8], fmt.Sprintf("%s/commit/%s", repoURL, line.Hash)))
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

	logger.Warnf("ignored markdown tag: %s", tag)

	return text
}

func tag(prefix, version string) string {
	return prefix + "v" + version
}

func link(text, repoURL string) string {
	return fmt.Sprintf("[%s](%s)", text, repoURL)
}

func writeTags(tagsFile string, tagPrefixes []string, newVers []semver) error {
	builder := &strings.Builder{}
	builder.WriteString("# this file is generated to aid in automating releases\n")

	for i, tagPrefix := range tagPrefixes {
		builder.WriteString(tag(tagPrefix, newVers[i].String()) + "\n")
	}

	return misc.Wrapfl(cmd.OverwriteFile(tagsFile, []byte(builder.String())))
}
