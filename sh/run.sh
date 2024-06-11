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

read -rd "$(printf \\r)" -a modules < <(
  findModulesPretty
  printf \\r
)

declare -ra commands=(
  "name:build               type:mod"
  "name:format              type:mod"
  "name:lint                type:mod"
  "name:lintFix             type:mod"
  # do not name "test"; shadowed by builtin test command
  "name:tests               type:mod"
  "name:ci                  type:repo"
  "name:ciPush              type:repo"
  "name:coverage            type:repo"
  "name:formatConfigs       type:repo"
  "name:generateMocks       type:repo"
  "name:spellcheckDocs      type:repo"
  "name:setup               type:repo"
  "name:tag                 type:repo"
  "name:testScripts         type:repo"
  "name:generateGoWork      type:repo"
  "name:newModule           type:repo"
  "name:generateVscodeTasks type:repo"
  "name:copyright           type:repo"
)

usageExit() {
  local regExpInvalidCapture='.nvalid\ [^:]+:\ ([[:alpha:]]*)' cmdNames=()

  msgln "$*"
  msgln usage:

  for info in "${commands[@]}"; do
    read -ra command <<<"$info"

    cmdNames+=("${command[0]/name:/}")

    msg "$0" "${command[0]/name:/}"

    if [ "${command[1]/type:/}" == mod ]; then
      printf "\t<module>"
    fi

    printf \\n
  done

  printf \\n
  msgln "modules: $(joinBy ', ' "${modules[@]}")"
  msgln use \'all\' as module to iterate all modules
  msgln pass -h to commands to see further options

  if [[ "$*" =~ $regExpInvalidCapture ]]; then
    didYouMean "${BASH_REMATCH[1]}" "${modules[@]}" "${cmdNames[@]}"
  fi

  exit 1
}

runCommandInModule() {
  local cmd=$1 module=$2

  run() {
    local cmd=$1 module=$2

    if ! $cmd "$module"; then
      msgln "failed: $cmd $module"
    fi

    if [ -d "$PWD/$module/${module}_test" ]; then
      if ! $cmd "$module/${module}_test"; then
        msgln "failed: $cmd $module/${module}_test"
      fi
    fi
  }

  if [ "$module" != all ]; then
    run $cmd "$module"
    return
  fi

  for mod in "${modules[@]}"; do
    printf \\n
    msgln "$cmd $mod..."
    run "$cmd" "$mod"
  done
}

# shellcheck disable=SC2317 # dynamic call
lint() {
  local path="$1" lintFlags=(--timeout 10s --print-issued-lines=false)

  golangci-lint run "${lintFlags[@]}" "$path"
}

# shellcheck disable=SC2317 # dynamic call
lintFix() {
  local path="$1"

  # lint fix is different from the linter
  ./sh/lint-fix.sh "$path" &
  backgroundLinter=$!

  local lintFlags=(--timeout 10s --print-issued-lines=false --fix)
  golangci-lint run "${lintFlags[@]}" "$path"
  wait "$backgroundLinter"
}

prettierFileGlob="**/*{.yml,.yaml,.json}"

# shellcheck disable=SC2317 # dynamic call
format() {
  local path="$1"

  gofumpt -l -w "$path"
  prettier --write "$1/$prettierFileGlob" 2>/dev/null || true
}

# shellcheck disable=SC2317 # dynamic call
formatConfigs() {
  prettier --write "./$prettierFileGlob" 2>/dev/null || true
}

# shellcheck disable=SC2317 # dynamic call
tests() {
  displayCoverage=""

  if [ "${inputArgs[1]:-}" == -cover ]; then
    displayCoverage=true
  fi

  MOD_PATH="$1" \
    CACHE="true" \
    DISPLAY_COVERAGE="$displayCoverage" \
    GITHUB_OUTPUT="/dev/null" \
    ./sh/workflows/module-pr/test-pretty.sh
}

# shellcheck disable=SC2317 # dynamic call
build() {
  install=""

  if [ "${inputArgs[1]:-}" == -install ]; then
    install=true
  fi

  MOD_PATH="$1" \
    INSTALL="$install" \
    ./sh/workflows/module-pr/build.sh
}

# shellcheck disable=SC2317 # dynamic call
ci() {
  requireGitClean
  requireInternet Internet required to pull docker images
  ./sh/ci.sh
}

# shellcheck disable=SC2317 # dynamic call
spellcheckDocs() {
  cspell "**/*.md" --gitignore
}

# shellcheck disable=SC2317 # dynamic call
setup() {
  ./sh/setup.sh
}

# shellcheck disable=SC2317 # dynamic call
testScripts() {
  find sh/sh_test -iname "*-test.sh" -exec ./{} \;
}

# shellcheck disable=SC2317 # dynamic call
tag() {
  # requireGitBranch main
  ./sh/tag.sh "$@"
}

# shellcheck disable=SC2317 # dynamic call
generateMocks() {
  ./sh/generate-mocks.sh
}

# shellcheck disable=SC2317 # dynamic call
coverage() {
  ./sh/coverage.sh
}

# shellcheck disable=SC2317 # dynamic call
generateGoWork() {
  go run cmd/gengowork/main.go
}

# shellcheck disable=SC2317 # dynamic call
newModule() {
  ./sh/new-module.sh "$@"
}

# shellcheck disable=SC2317 # dynamic call
generateVscodeTasks() {
  go run cmd/genvscodetasks/main.go
}

# shellcheck disable=SC2317 # dynamic call
ciPush() {
  requireGitClean
  requireInternet Internet required to pull docker images
  ./sh/ci.sh push
}

# shellcheck disable=SC2317 # dynamic call
copyright() {
  go run cmd/copyright/main.go "$@"
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "a command is required"
fi

inputCommand=${1}
declare -a inputArgs=("${@:2}")

### script ###

for info in "${commands[@]}"; do
  read -ra command <<<"$info"

  if [ "$inputCommand" != "${command[0]/name:/}" ]; then
    continue
  fi

  if requestedHelp "$*"; then
    # forward -h to the command
    "${command[0]/name:/}" "${inputArgs[@]}"
    exit 1
  fi

  if [ "${command[1]/type:/}" == mod ]; then
    if ! [[ " all ${modules[*]} " =~ ${inputArgs[0]} ]]; then
      usageExit "invalid module: ${inputArgs[0]}"
    fi

    runCommandInModule "${command[0]/name:/}" "${inputArgs[@]}"
    exit
  fi

  "${command[0]/name:/}" "${inputArgs[@]}"
  exit
done

usageExit "invalid command: ${inputCommand}"
