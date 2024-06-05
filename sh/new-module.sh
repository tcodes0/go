#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

usageExit() {
  msgln "Usage: $0 <name>"
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

msgln "todo:
  - run script to update go.work
  - add $module to github workflows
  - add $module to vscode json config
"
