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

MAIN_WORKFLOW=.github/workflows/main.yml
CHANGED_FILES=.changed-files.yml

usage() {
  command cat <<-EOF
Usage:
Create a new command

$0 pizza
create a new cmd called t0pizza
EOF
}

init() {
  local name="t0$1" go_ver

  read -r _ _ go_ver _ < <(go version)

  local format="
  cmd-%s:
    name: cmd/%s
    needs: changed-files
    if: needs.changed-files.outputs.TODO
    uses: ./.github/workflows/module_pr.yml
    with:
      goVersion: ${go_ver/go/}
      modulePath: cmd/%s
"

  command cp -RH cmd/template "cmd/$name"
  echo "cmd/$name"
  # shellcheck disable=SC2059 # format variable
  printf "$format" "$name" "$name" "$name" >>$MAIN_WORKFLOW
  printf "cmd_%s:\n  - cmd/%s/**.go\n" "$name" "$name" >>$CHANGED_FILES
}

cleanup() {
  go run cmd/gengowork/main.go
  go run cmd/t0copyright/main.go -check "*.go" -fix -comment '// '

  msgln "next steps:
  - edit $MAIN_WORKFLOW and add output variable to changed-files step
  - update new module job with output variable (currently TODO)
  - review $CHANGED_FILES"
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
