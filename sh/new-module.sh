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

name="${1:-}"
formatMod="
  %s:
    name: %s
    needs: changed-files
    if: needs.changed-files.outputs.TODO
    uses: ./.github/workflows/module-pr.yml
    with:
      goVersion: TODO
      modulePath: %s
"
formatCmd="
  cmd-%s:
    name: cmd/%s
    needs: changed-files
    if: needs.changed-files.outputs.TODO
    uses: ./.github/workflows/module-pr.yml
    with:
      goVersion: TODO
      modulePath: cmd/%s
"

### validation, input handling ###

if requestedHelp "$*" || [ -z "$name" ]; then
  msgln "Inputs:"
  msgln "<name>\t initializes a new go module called <name>\t (required)"
  msgln "-cmd\t instead of a module init cmd/<name>"
  exit 1
fi

### script ###

module="github.com/tcodes0/go/$name"
format=""

if [[ "$*" =~ -cmd ]]; then
  mkdir -p "cmd/$name"
  cd "cmd/$name"
  printf "package main\n" >"main.go"
  touch config.yml
  format=$formatCmd
else
  mkdir -p "$name/${name}_test"
  cd "$name"
  go mod init "$module"
  printf "package %s\n" "$name" >"$name.go"
  format=$formatMod
fi

\cd -

# shellcheck disable=SC2059 # format variable
printf "$format" "$name" "$name" "$name" >>.github/workflows/main.yml

msgln "todo:
  - ./run <generate go work task>
  - ./run <generate vscode task config task>
  - ./run <copyright task>
  - edit github workflows"
