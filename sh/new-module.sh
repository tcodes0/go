#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

usageExit() {
  msgLn "Usage: $0 <name>"
  exit 1
}

name="${1:-}"

### validation, input handling ###

if requestedHelp "$*"; then
  usageExit
fi

if [ -z "$name" ]; then
  usageExit
fi

### script ###

module="$ROOT_MODULE/$name"

mkdir -p "$name/${name}_test"
cd "$name"
go mod init "$module"
printf "package %s\n" "$name" >"$name.go"

msgLn "todo:
  - run script to update go.work
  - add $module to github workflows
  - add $module to vscode json config
"
