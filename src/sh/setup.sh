#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

pass() {
  printf "%b\n" "$PASS_GREEN $1"
}

fail() {
  printf "%b\n" "$FAIL_RED $1"
}

installCommands=()

# $1 - package, $2 - binary, $* - comment
verifyGoTool() {
  pkg="$1"
  shift
  binary="$1"
  shift
  comment="$*"

  if ! command -v "$binary" >/dev/null; then
    fail "$binary $comment"
    installCommands+=("go install $pkg")
  else
    pass "$binary"
  fi
}

# $1 - package, $2 - binary, $* - comment
verifyJSTool() {
  pkg="$1"
  shift
  binary="$1"
  shift
  comment="$*"

  if ! command -v "$binary" >/dev/null; then
    fail "$binary $comment"
    installCommands+=("npm install -g  $pkg")
  else
    pass "$binary"
  fi
}

# $1 - install Link, $2 - binary, $* - comment
verifyManualTool() {
  link="$1"
  shift
  binary="$1"
  shift
  comment="$*"

  if ! command -v "$binary" >/dev/null; then
    fail "$binary $comment"
    installCommands+=("echo $link")
  else
    pass "$binary"
  fi
}

exitWithIssues() {
  if [ ${#installCommands[@]} -gt 0 ]; then
    printf "\n"
    msg "$1"
    for cmd in "${installCommands[@]}"; do
      printf '%s\n' "$cmd"
    done
    exit 1
  fi
}

# by order of priority
# basic gnu/linux tools included by default, git, etc...
verifyManualTool 'missing git' git a version control system
verifyManualTool 'missing bash' bash popular shell
verifyManualTool 'missing sed' sed stream editor
verifyManualTool 'missing mktemp' mktemp create temporary files and directories
verifyManualTool 'missing tput' tput terminal control
verifyManualTool 'missing find' find search for files in a directory hierarchy
verifyManualTool 'missing wc' wc word, line, character, and byte count
verifyManualTool 'missing date' date display the system date and time
verifyManualTool 'missing sort' sort basic sorting program
verifyManualTool 'missing uniq' uniq removes duplicates from input
verifyManualTool 'missing tr' tr translates characters
verifyManualTool 'missing tee' tee pipe input to two programs
verifyManualTool 'missing ps' ps view running programs
verifyManualTool 'missing grep' grep search files for matches
verifyManualTool 'missing sleep' sleep block a script for some time
verifyManualTool 'missing head' head read a number of lines from a file

exitWithIssues "missing basic gnu/linux binaries; please install for your platform, good luck!"

# programming languages, package managers
verifyManualTool 'see https://nodejs.org/en/download/package-manager/' node javascript runtime built on top of v8
verifyManualTool 'see https://go.dev/doc/install' go a static, compiled, minimalistic, garbage collected language

exitWithIssues "install the programming languages then run this script again"

# Go
verifyGoTool mvdan.cc/gofumpt@latest gofumpt is a stricter gofmt
verifyGoTool github.com/go-delve/delve/cmd/dlv@latest dlv delve go debugger
verifyGoTool github.com/joho/godotenv/cmd/godotenv@latest godotenv runs go programs with a .env local file
verifyGoTool github.com/fatih/gomodifytags@latest gomodifytags is a tool to modify struct field tags easily
verifyGoTool github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest gopkgs a faster go list all
verifyGoTool golang.org/x/tools/gopls@latest gopls go language server
verifyGoTool github.com/cweill/gotests/gotests@latest gotests a test generator
verifyGoTool github.com/vektra/mockery/v2@latest mockery a mock interface generator
verifyGoTool github.com/ramya-rao-a/go-outline@latest go-outline utility to extract a json representation of a go source file
verifyGoTool github.com/haya14busa/goplay/cmd/goplay@latest goplay playground client of https://play.golang.org
verifyGoTool github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest gotestfmt go test output formatter
verifyGoTool github.com/josharian/impl@latest impl generates method stubs for implementing an interface
verifyGoTool honnef.co/go/tools/cmd/staticcheck@latest staticcheck a go mega linter
verifyGoTool mvdan.cc/sh/v3/cmd/shfmt@latest shfmt formats shell scripts

# Official auto fix tools for go vet linter
verifyGoTool golang.org/x/tools/go/analysis/passes/defers/cmd/defers@latest defers
verifyGoTool golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest fieldalignment
verifyGoTool golang.org/x/tools/go/analysis/passes/findcall/cmd/findcall@latest findcall
verifyGoTool golang.org/x/tools/go/analysis/passes/httpmux/cmd/httpmux@latest httpmux
verifyGoTool golang.org/x/tools/go/analysis/passes/ifaceassert/cmd/ifaceassert@latest ifaceassert
verifyGoTool golang.org/x/tools/go/analysis/passes/lostcancel/cmd/lostcancel@latest lostcancel
verifyGoTool golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness@latest nilness
verifyGoTool golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest shadow
verifyGoTool golang.org/x/tools/go/analysis/passes/stringintconv/cmd/stringintconv@latest stringintconv
verifyGoTool golang.org/x/tools/go/analysis/passes/unmarshal/cmd/unmarshal@latest unmarshal
verifyGoTool golang.org/x/tools/go/analysis/passes/unusedresult/cmd/unusedresult@latest unusedresult

# JS
verifyJSTool cspell@latest cspell a spellchecker for source code
verifyJSTool prettier@latest prettier a code formatter for several languages

# others
verifyManualTool 'see https://golangci-lint.run/welcome/install' golangci-lint a fast lint runner for Go
verifyManualTool 'see https://vektra.github.io/mockery/latest/installation' mockery a go code generator for tests
verifyManualTool 'see https://nektosact.com/installation/index.html' act run github actions locally using containers
verifyManualTool 'see https://github.com/cli/cli#installation' gh new github CLI

exitWithIssues "install the missing tools with"

# configuration
if ! gh auth token >/dev/null 2>&1; then
  fail 'gh auth token' 'not signed in to gh'
  installCommands+=("please sign in to gh using 'gh auth login'")
else
  pass 'gh auth token'
fi

msg after installing act, run \'act\' to setup

exitWithIssues "fix configuration issues"
