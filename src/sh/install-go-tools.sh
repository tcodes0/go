#! /usr/bin/env bash

set -euo pipefail

echoInstall() {
  pkg="$1"
  shift
  echo "$@"
  GOPATH=~/Desktop/gopath go install "$pkg"
  echo
}

# dlv godotenv gomodifytags gopkgs gopls gotests mockery dlv-dap go-outline goplay gotestfmt impl staticcheck
# plus autofixers from govet

echoInstall mvdan.cc/gofumpt@latest gofumpt is a stricter gofmt
echoInstall github.com/go-delve/delve/cmd/dlv@latest delve go debugger
echoInstall github.com/joho/godotenv/cmd/godotenv@latest godotenv runs go programs with a .env local file
echoInstall github.com/fatih/gomodifytags@latest gomodifytags is a tool to modify struct field tags easily
echoInstall github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest gopkgs a faster go list all
echoInstall golang.org/x/tools/gopls@latest gopls go language server
echoInstall github.com/cweill/gotests/gotests@latest gotests a test generator
echoInstall github.com/vektra/mockery/v2@latest mockery a mock interface generator
echoInstall github.com/ramya-rao-a/go-outline@latest go-outline utility to extract a json representation of a go source file
echoInstall github.com/haya14busa/goplay/cmd/goplay@latest goplay playground client of https://play.golang.org
echoInstall github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest gotestfmt go test output formatter
echoInstall github.com/josharian/impl@latest impl generates method stubs for implementing an interface
echoInstall honnef.co/go/tools/cmd/staticcheck@latest staticcheck a go mega linter

if ! command -v golangci-lint >/dev/null; then
  echo "golangci-lint is a fast lint runner for Go"
  echo "see https://golangci-lint.run/welcome/install for instructions"
  echo "or install via package manager"
fi
