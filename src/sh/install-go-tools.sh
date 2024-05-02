#! /usr/bin/env bash

set -euo pipefail

alreadyInstalled() {
  echo -e "$1\t\t is already installed"
}

# $1 - package, $2 - binary, $3 - comment
echoInstall() {
  pkg="$1"
  shift
  binary="$1"
  shift
  comment="$*"

  if ! command -v "$binary" >/dev/null; then
    if [ -n "$comment" ]; then
      echo "$comment"
    fi

    echo go install "$pkg"
    echo
  else
    alreadyInstalled "$binary"
  fi
}

echo -e "\n\topen source tools"
echoInstall mvdan.cc/gofumpt@latest gofumpt gofumpt is a stricter gofmt
echoInstall github.com/go-delve/delve/cmd/dlv@latest dlv delve go debugger
echoInstall github.com/joho/godotenv/cmd/godotenv@latest godotenv godotenv runs go programs with a .env local file
echoInstall github.com/fatih/gomodifytags@latest gomodifytags gomodifytags is a tool to modify struct field tags easily
echoInstall github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest gopkgs gopkgs a faster go list all
echoInstall golang.org/x/tools/gopls@latest gopls gopls go language server
echoInstall github.com/cweill/gotests/gotests@latest gotests gotests a test generator
echoInstall github.com/vektra/mockery/v2@latest mockery mockery a mock interface generator
echoInstall github.com/ramya-rao-a/go-outline@latest go-outline go-outline utility to extract a json representation of a go source file
echoInstall github.com/haya14busa/goplay/cmd/goplay@latest goplay goplay playground client of https://play.golang.org
echoInstall github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest gotestfmt gotestfmt go test output formatter
echoInstall github.com/josharian/impl@latest impl impl generates method stubs for implementing an interface
echoInstall honnef.co/go/tools/cmd/staticcheck@latest staticcheck staticcheck a go mega linter

if ! command -v golangci-lint >/dev/null; then
  echo "golangci-lint is a fast lint runner for Go"
  echo "see https://golangci-lint.run/welcome/install for instructions"
  echo "or install via package manager"
else
  alreadyInstalled golangci-lint
fi

echo -e "\n\tofficial auto fix go vet tools"
echoInstall golang.org/x/tools/go/analysis/passes/defers/cmd/defers@latest defers
echoInstall golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest fieldalignment
echoInstall golang.org/x/tools/go/analysis/passes/findcall/cmd/findcall@latest findcall
echoInstall golang.org/x/tools/go/analysis/passes/httpmux/cmd/httpmux@latest httpmux
echoInstall golang.org/x/tools/go/analysis/passes/ifaceassert/cmd/ifaceassert@latest ifaceassert
echoInstall golang.org/x/tools/go/analysis/passes/lostcancel/cmd/lostcancel@latest lostcancel
echoInstall golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness@latest nilness
echoInstall golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest shadow
echoInstall golang.org/x/tools/go/analysis/passes/stringintconv/cmd/stringintconv@latest stringintconv
echoInstall golang.org/x/tools/go/analysis/passes/unmarshal/cmd/unmarshal@latest unmarshal
echoInstall golang.org/x/tools/go/analysis/passes/unusedresult/cmd/unusedresult@latest unusedresult
