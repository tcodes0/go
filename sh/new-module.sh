#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

module="${1:-}"

### validation, input handling ###

if [ -z "$module" ]; then
  echo "Usage: $0 <module>"
  exit 1
fi

### script ###

mkdir -p "$module/${module}_test"
cd "$module"
go mod init "$module"
printf "package %s\n" "$module" >"$module.go"

msg "todo:
  - run script to update go.work
  - add $module to github workflows
  - add $module to vscode json config
"
