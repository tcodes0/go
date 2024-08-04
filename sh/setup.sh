#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck source=lib.sh
source "$PWD/sh/lib.sh"

fixProblems=()

pass() {
  printf "$LIB_COLOR_PASS$LIB_FORMAT_DIM %b$LIB_VISUAL_END\n" "$1"
}

fail() {
  printf "$LIB_COLOR_FAIL %b$LIB_VISUAL_END\n" "$1"
}

setup() {
  type="$1"
  installLink="$2"
  binary="$3"
  comments="${*:4}"
  declare -A installCommandsByType=(
    ["go"]="go install"
    ["js"]="npm install --global"
    ["manual"]="-"
  )

  if ! command -v "$binary" >/dev/null; then
    fail "$binary $comments"
    fixProblems+=("${installCommandsByType[$type]} $installLink")
  else
    pass "$binary"
  fi
}

exitShowProblems() {
  if [ ${#fixProblems[@]} == 0 ]; then
    return
  fi

  printf \\n
  msgln "$1"

  for cmd in "${fixProblems[@]}"; do
    printf '%s\n' "$cmd"
  done

  exit 1
}

if requestedHelp "$*"; then
  msgln "check for missing tools, configuration and show notes"
  exit 1
fi

if [ "$(basename "$PWD")" != "tcodes0-go" ]; then
  msgln "run this script from project root"
  exit 1
fi

# by order of priority

# basic gnu/linux tools included by default, git, etc...

setup manual 'missing git' git a version control system
setup manual 'missing bash' bash popular shell
setup manual 'missing mktemp' mktemp create temporary files and directories
setup manual 'missing tput' tput terminal control
setup manual 'missing find' find search for files in a directory hierarchy
setup manual 'missing wc' wc word, line, character, and byte count
setup manual 'missing date' date display the system date and time
setup manual 'missing sort' sort basic sorting program
setup manual 'missing uniq' uniq removes duplicates from input
setup manual 'missing tr' tr translates characters
setup manual 'missing tee' tee pipe input to two programs
setup manual 'missing ps' ps view running programs
setup manual 'missing grep' grep search files for matches
setup manual 'missing sleep' sleep block a script for some time
setup manual 'missing head' head read a number of lines from a file
setup manual 'missing less' less pager to view files
setup manual 'missing tail' tail read the end of a file
setup manual 'missing uname' uname print system information

if macos; then
  setup manual 'missing gsed' gsed gnu sed stream editor, available on brew as 'gnu-sed'
else
  setup manual 'missing sed' sed stream editor
fi

exitShowProblems "missing basic gnu/linux binaries; please install for your platform; seek help and good luck!"

# programming languages, package managers

setup manual 'see https://github.com/nvm-sh/nvm?tab=readme-ov-file#installing-and-updating then run "nvm install node"' node javascript runtime built on top of v8
setup manual 'see https://go.dev/doc/install' go a static, compiled, minimalistic, garbage collected language

exitShowProblems "install the programming languages then run this script again"

# go tools

setup go mvdan.cc/gofumpt@latest gofumpt is a stricter gofmt
setup go github.com/go-delve/delve/cmd/dlv@latest dlv delve go debugger
setup go github.com/joho/godotenv/cmd/godotenv@latest godotenv runs go programs with a .env local file
setup go github.com/fatih/gomodifytags@latest gomodifytags is a tool to modify struct field tags easily
setup go github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest gopkgs a faster go list all
setup go golang.org/x/tools/gopls@latest gopls go language server
setup go golang.org/x/tools/cmd/cover@latest cover go coverage tool
setup go github.com/cweill/gotests/gotests@latest gotests a test generator
setup go github.com/ramya-rao-a/go-outline@latest go-outline utility to extract a json representation of a go source file
setup go github.com/haya14busa/goplay/cmd/goplay@latest goplay playground client of https://play.golang.org
setup go github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest gotestfmt go test output formatter
setup go github.com/josharian/impl@latest impl generates method stubs for implementing an interface
setup go honnef.co/go/tools/cmd/staticcheck@latest staticcheck a go mega linter
setup go mvdan.cc/sh/v3/cmd/shfmt@latest shfmt formats shell scripts
setup go golang.org/x/tools/cmd/goimports@latest goimports updates go import lines

# auto fix tools for go vet linter

setup go golang.org/x/tools/go/analysis/passes/defers/cmd/defers@latest defers
setup go golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest fieldalignment
setup go golang.org/x/tools/go/analysis/passes/findcall/cmd/findcall@latest findcall
setup go golang.org/x/tools/go/analysis/passes/httpmux/cmd/httpmux@latest httpmux
setup go golang.org/x/tools/go/analysis/passes/ifaceassert/cmd/ifaceassert@latest ifaceassert
setup go golang.org/x/tools/go/analysis/passes/lostcancel/cmd/lostcancel@latest lostcancel
setup go golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness@latest nilness
setup go golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest shadow
setup go golang.org/x/tools/go/analysis/passes/stringintconv/cmd/stringintconv@latest stringintconv
setup go golang.org/x/tools/go/analysis/passes/unmarshal/cmd/unmarshal@latest unmarshal
setup go golang.org/x/tools/go/analysis/passes/unusedresult/cmd/unusedresult@latest unusedresult
setup go github.com/4meepo/tagalign/cmd/tagalign@latest tagalign

# JS

setup js cspell@latest cspell a spellchecker for source code
setup js prettier@latest prettier a code formatter for several languages
setup js @commitlint/cli@latest commitlint a linter for commit messages

# others

setup manual 'see https://golangci-lint.run/welcome/install' golangci-lint a fast lint runner for Go
setup manual 'see https://vektra.github.io/mockery/latest/installation' mockery a go code generator for tests
setup manual 'see https://nektosact.com/installation/index.html' act run github actions locally using containers
setup manual 'see https://github.com/cli/cli#installation' gh new github CLI
setup manual 'see https://github.com/koalaman/shellcheck?tab=readme-ov-file#installing' shellcheck shell script linter
setup manual 'see https://docs.docker.com/get-docker/' docker container runtime

exitShowProblems "install the missing tools with"

# configuration

if ! [[ "$SHELL" =~ bash ]]; then
  fail 'shell is bash' "expected bash as shell but got $SHELL"
  fixProblems+=("either use bash as default shell or start bash as subshell using 'bash'")
else
  pass 'shell is bash'
fi

if ! gh auth token >/dev/null 2>&1; then
  fail 'gh auth token' 'not signed in to gh'
  fixProblems+=("please sign in to gh using 'gh auth login'")
else
  pass 'gh auth token'
fi

if ! docker stats --no-stream >/dev/null 2>&1; then
  fail 'docker running' 'docker daemon not running'
  fixProblems+=("please start docker")
else
  pass 'docker running'
fi

if ! MOD_PATH=cmd/runner ./sh/workflows/module-pr/build.sh; then
  fail 'build cmd/runner' 'build failed'
  fixProblems+=("cmd/runner build failed, ./run symlink won't work")
else
  pass 'build cmd/runner'
fi

if [ ! -d wiki/.git ]; then
  fail 'local wiki' 'wiki is not cloned'
  fixProblems+=("git clone git@github.com:tcodes0/go.wiki.git wiki")
else
  pass 'local wiki'
fi

if [ ! -f .env ]; then
  fail '.env' '.env not copied'
  fixProblems+=("cp .env-default .env")
else
  pass '.env'
fi

exitShowProblems "fix configuration issues:"

# notes

msgln
msgln note: before using ./ci, run \'act\' once to set it up
msgln note: run \'export CMD_COLOR=true\' to see colored output, or add to .env

if [ "${NVM_DIR:-}" ]; then
  msgln note: when using nvm and upgrading node, global packages need to be reinstalled
fi
