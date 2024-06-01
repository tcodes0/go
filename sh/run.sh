#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

declare -rA opts=(
  [all]="all"
)

read -rd "$CHAR_CARRIG_RET" -a modules < <(
  regExpDotSlashPrefix="^\.\/"
  # find folders directly under . that have at least 1 *.go file; prettify output
  findModules | sed -e "s/$regExpDotSlashPrefix//" | tr '\n' ' '

  printf %b "$CHAR_CARRIG_RET"
)

declare -ra modulesAndAll=("${opts[all]}" "${modules[@]}")

declare -ra commands=(
  "name:build          type:mod  argCount:1"
  "name:format         type:mod  argCount:1"
  "name:lint           type:mod  argCount:1"
  "name:lintFix        type:mod  argCount:1"
  # do not name "test"; shadowed builtin test command
  "name:tests          type:mod  argCount:1"
  "name:ci             type:repo argCount:0"
  "name:coverage       type:repo argCount:0"
  "name:formatConfigs  type:repo argCount:0"
  "name:generateMocks  type:repo argCount:0"
  "name:spellcheckDocs type:repo argCount:0"
  "name:setup          type:repo argCount:0"
  "name:tag            type:repo argCount:2"
  "name:testScripts    type:repo argCount:0"
  "name:generateGoWork type:repo argCount:0"
  "name:newModule      type:repo argCount:1"
)

usageExit() {
  usage() {
    name=$1
    argCount=$2

    msg "$0" "$name"

    if [ "$argCount" != 0 ]; then
      printf %b \\t

      for ((i = 1; i <= argCount; i++)); do
        printf "<arg%s>\t" $i
      done
    fi

    printf %b \\n
  }

  msgLn "$*\n"
  msgLn Usage:

  for info in "${commands[@]}"; do
    read -ra cmdInfo <<<"$info"

    usage "${cmdInfo[0]/name:/}" "${cmdInfo[2]/argCount:/}"
  done

  printf %b \\n
  msgLn "modules:\n$(joinBy ', ' "${modulesAndAll[@]}")"

  exit 1
}

# shellcheck disable=SC2317 # dynamic call
lint() {
  local path="$1"
  local lintFlags=(--timeout 10s --print-issued-lines=false)

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
  MOD_PATH="$1" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./sh/workflows/module-pr/test-pretty.sh
}

# shellcheck disable=SC2317 # dynamic call
build() {
  MOD_PATH="$1" \
    ./sh/workflows/module-pr/build-go.sh && echo ok
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
  requireGitBranch main
  ./sh/tag.sh "$@"
}

run() {
  local command=$1
  local module=$2

  $command "$module" || true

  if [ -d "$PWD/$module/${module}_test" ]; then
    $command "$module/${module}_test" || true
  fi
}

runCommandInModule() {
  local command=$1
  local module=$2

  if [ "$module" != "${opts[all]}" ]; then
    run $command "$module"
    return
  fi

  for mod in "${modules[@]}"; do
    printf %b "\n"
    msgLn "$command $mod..."
    run "$command" "$mod"
  done
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
  ./sh/generate-go-work.sh
}

# shellcheck disable=SC2317 # dynamic call
newModule() {
  ./sh/new-module.sh "$@"
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "A command is required"
fi

inputCommand=${1}
declare -a inputArgs=("${@:2}")

### script ###

for info in "${commands[@]}"; do
  read -ra command <<<"$info"

  if [ "$inputCommand" != "${command[0]/name:/}" ]; then
    continue
  fi

  if [ "${#inputArgs[@]}" != "${command[2]/argCount:/}" ]; then
    usageExit "${command[0]/name:/} wants ${command[2]/argCount:/} arguments; received ${#inputArgs[@]} (${inputArgs[*]})"
  fi

  if [ "${command[1]/type:/}" == mod ]; then
    if ! [[ " ${modulesAndAll[*]} " =~ ${inputArgs[0]} ]]; then
      usageExit "Invalid module: ${inputArgs[0]}"
    fi

    runCommandInModule "${command[0]/name:/}" "${inputArgs[@]}"
    exit
  fi

  "${command[0]/name:/}" "${inputArgs[@]}"
  exit
done

usageExit "Invalid command: ${inputCommand}"
