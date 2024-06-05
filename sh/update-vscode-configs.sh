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

vscodeRoot=.vscode
tasksFile=tasks.json
inputModCmds="${1:-}"
inputRepoCmds="${2:-}"

read -ra modCmds <<<"$inputModCmds"
read -ra repoCmds <<<"$inputRepoCmds"

usageExit() {
  msgln "$*"
  msgln "usage: $0 '<module commands>' '<repo commands>'"
  exit 1
}

setArray() {
  local file="$1" expr="$2" values="${*:3}" jsonValues=() joined=""

  for value in $values; do
    jsonValues+=("$(printf '"%s"' "$value")")
  done

  joined=$(joinBy , "${jsonValues[@]}")

  yq --inplace eval "$expr = [$joined]" "$file"
}

### validation, input handling ###

if ! [ -d "$vscodeRoot" ]; then
  msgExit "vscode root not found: $vscodeRoot"
fi

if ! [ -f "$vscodeRoot/$tasksFile" ]; then
  msgExit "file not found: $tasksFile"
fi

if [ -z "$inputModCmds" ]; then
  usageExit "missing module commands"
fi

if [ -z "$inputRepoCmds" ]; then
  usageExit "missing repo commands"
fi

### script ###

setArray "$vscodeRoot/$tasksFile" .inputs[0].options "all $(findModulesPretty)"
setArray "$vscodeRoot/$tasksFile" .inputs[1].options "${modCmds[@]}"
setArray "$vscodeRoot/$tasksFile" .inputs[2].options "${repoCmds[@]}"

prettier --write "$vscodeRoot/$tasksFile" 2>/dev/null
