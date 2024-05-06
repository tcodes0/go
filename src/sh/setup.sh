#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

import() {
  relativePath="go\/src\/sh\/lib.sh"
  regExpBasePath="(.*)\/go\/?.*"
  functions=$(sed -E "s/$regExpBasePath/\1\/$relativePath/g" <<<"$PWD")

  # shellcheck disable=SC1090
  source "$functions"
}

import

pass() {
  printf "%b\n" "$PASS_GREEN $1"
}

fail() {
  printf "%b\n" "$FAIL_RED $1"
}

installCommands=()

# $1 - package, $2 - binary, $* - comment
verifyTool() {
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

############################################
msg "open source tools"
verifyTool mvdan.cc/gofumpt@latest gofumpt is a stricter gofmt
verifyTool github.com/go-delve/delve/cmd/dlv@latest dlv delve go debugger
verifyTool github.com/joho/godotenv/cmd/godotenv@latest godotenv runs go programs with a .env local file
verifyTool github.com/fatih/gomodifytags@latest gomodifytags is a tool to modify struct field tags easily
verifyTool github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest gopkgs a faster go list all
verifyTool golang.org/x/tools/gopls@latest gopls go language server
verifyTool github.com/cweill/gotests/gotests@latest gotests a test generator
verifyTool github.com/vektra/mockery/v2@latest mockery a mock interface generator
verifyTool github.com/ramya-rao-a/go-outline@latest go-outline utility to extract a json representation of a go source file
verifyTool github.com/haya14busa/goplay/cmd/goplay@latest goplay playground client of https://play.golang.org
verifyTool github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest gotestfmt go test output formatter
verifyTool github.com/josharian/impl@latest impl generates method stubs for implementing an interface
verifyTool honnef.co/go/tools/cmd/staticcheck@latest staticcheck a go mega linter
verifyManualTool 'see https://golangci-lint.run/welcome/install' golangci-lint a fast lint runner for Go

###########################################
printf "\n"
msg "official auto fix tools for go vet linter"
verifyTool golang.org/x/tools/go/analysis/passes/defers/cmd/defers@latest defers
verifyTool golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest fieldalignment
verifyTool golang.org/x/tools/go/analysis/passes/findcall/cmd/findcall@latest findcall
verifyTool golang.org/x/tools/go/analysis/passes/httpmux/cmd/httpmux@latest httpmux
verifyTool golang.org/x/tools/go/analysis/passes/ifaceassert/cmd/ifaceassert@latest ifaceassert
verifyTool golang.org/x/tools/go/analysis/passes/lostcancel/cmd/lostcancel@latest lostcancel
verifyTool golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness@latest nilness
verifyTool golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest shadow
verifyTool golang.org/x/tools/go/analysis/passes/stringintconv/cmd/stringintconv@latest stringintconv
verifyTool golang.org/x/tools/go/analysis/passes/unmarshal/cmd/unmarshal@latest unmarshal
verifyTool golang.org/x/tools/go/analysis/passes/unusedresult/cmd/unusedresult@latest unusedresult

###########################################

if [ ${#installCommands[@]} -gt 0 ]; then
  printf "\n"
  msg "install the missing tools with"
  for cmd in "${installCommands[@]}"; do
    printf '%s\n' "$cmd"
  done
fi
