#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar nullglob
trap 'err $LINENO' ERR

##########################
### vars and functions ###
##########################

usage() {
  command cat <<-EOF
Usage:
Create a new module or new command

$0 pizza
create a new module called pizza

$0 pizza -cmd
create a new command called pizza
EOF
}

init() {
  local formatMod="
  %s:
    name: %s
    needs: changed-files
    if: needs.changed-files.outputs.TODO
    uses: ./.github/workflows/module_pr.yml
    with:
      goVersion: TODO
      modulePath: %s
"
  local formatCmd="
  cmd-%s:
    name: cmd/%s
    needs: changed-files
    if: needs.changed-files.outputs.TODO
    uses: ./.github/workflows/module_pr.yml
    with:
      goVersion: TODO
      modulePath: cmd/%s
"
  local name="$1" format files_entry
  local module="github.com/tcodes0/go/$name"

  if [[ "$*" =~ -cmd ]]; then
    command cp -RH cmd/template "cmd/$name"
    format=$formatCmd
  else
    command mkdir -p "$name/${name}_test"
    command cd "$name"
    go mod init "$module"
    printf "package %s\n" "$name" >"$name.go"
    format=$formatMod
    command cd -
  fi

  # shellcheck disable=SC2059 # format variable
  printf "$format" "$name" "$name" "$name" >>.github/workflows/main.yml

  if [[ "$*" =~ -cmd ]]; then
    files_entry=$name
    printf "cmd_%s:\n  - cmd/%s/**.go\n" "$name" "$name" >>.github/workflows/files.yml
  else
    files_entry="cmd_$name"
    printf "%s:\n  - %s/**.go\n  - go.mod\n  - go.sum\n" "$name" "$name" >>.github/workflows/main.yml
  fi
}

cleanup() {
  go run cmd/gengowork/main.go
  go run cmd/t0copyright/main.go -fix -find "*.go" -comment '// '

  msgln "todo:
  - edit .github/workflows/main.yml to fix TODOs and add $files_entry output"
}

##############
### script ###
##############

if requested_help "$*" || [ ! "${1:-}" ]; then
  usage
  exit 1
fi

init "$1"
cleanup
