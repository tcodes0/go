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

declare -rA moduleCommands=(
  [build]="build"
  [format]="format"
  [lint]="lint"
  [lintfix]="lint-fix"
  [test]="test"
)

declare -rA repoCommands=(
  [ci]="ci"
  [coverage]="coverage"
  [formatConfigs]="format-configs"
  [mocks]="generate-mocks"
  [spellcheck]="spellcheck-docs"
  [setup]="setup"
  [tag]="tag"
  [testSh]="test-scripts"
  [goWork]="generate-go-work"
  [newMod]="new-module"
)

declare -rA repoCommandArgs=(
  [ci]="0"
  [coverage]="0"
  [formatConfigs]="0"
  [mocks]="0"
  [spellcheck]="0"
  [setup]="0"
  [tag]="2"
  [testSh]="0"
  [goWork]="0"
  [newMod]="1"
)

declare -A optValue=(
  # defaults
  [command]=""
  [module]=""
)

declare -A repoCommandValues=()

for key in "${!repoCommands[@]}"; do
  repoCommandValues["${repoCommands[$key]}"]=$key
done

usageExit() {
  msg "$*\n"
  msg "Usage: $0 <repo command>"
  msg "Usage: $0 <module command> <module>"
  msg "repo commands:\n\t$(joinBy '\n\t' "${repoCommands[@]}")"
  msg "module commands:\n\t$(joinBy '\n\t' "${moduleCommands[@]}")"
  msg "modules:\n\t$(joinBy '\n\t' "${modulesAndAll[@]}")"

  exit 1
}

lint() {
  local path="$1"
  local lintFlags=(--timeout 10s --print-issued-lines=false)

  golangci-lint run "${lintFlags[@]}" "$path"
}

lintFix() {
  local path="$1"

  # lint fix is different from the linter
  ./sh/lint-fix.sh "$path" &
  backgroundLinter=$!

  local lintFlags=(--timeout 10s --print-issued-lines=false --fix)
  golangci-lint run "${lintFlags[@]}" "$path"
  wait $backgroundLinter
}

prettierFileGlob="**/*{.yml,.yaml,.json}"

format() {
  local path="$1"

  gofumpt -l -w "$path"
  prettier --write "$1/$prettierFileGlob" 2>/dev/null || true
}

formatConfigs() {
  prettier --write "./$prettierFileGlob" 2>/dev/null || true
}

unitTests() {
  MOD_PATH="$1" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./sh/workflows/module-pr/test-pretty.sh
}

build() {
  MOD_PATH="$1" \
    ./sh/workflows/module-pr/build-go.sh && echo ok
}

ci() {
  requireGitClean
  requireInternet Internet required to pull docker images
  ./sh/ci.sh
}

spellcheckDocs() {
  cspell "**/*.md" --gitignore
}

setup() {
  ./sh/setup.sh
}

testScripts() {
  find sh/sh_test -iname "*-test.sh" -exec ./{} \;
}

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
    msg "$command $mod..."
    run "$command" "$mod"
  done
}

generateMocks() {
  ./sh/generate-mocks.sh
}

coverage() {
  ./sh/coverage.sh
}

goWork() {
  ./sh/generate-go-work.sh
}

newMod() {
  ./sh/new-module.sh "$@"
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "One or more arguments required"
fi

optValue[command]=$1
optValue[module]=${2:-}

if ! [[ " ${moduleCommands[*]}${repoCommands[*]} " =~ ${optValue[command]} ]]; then
  usageExit "Invalid command: ${optValue[command]}"
fi

if [[ " ${moduleCommands[*]} " =~ ${optValue[command]} ]]; then
  if [ -z "${optValue[module]}" ]; then
    usageExit "Command ${optValue[command]} requires a module"
  fi

  if ! [[ " ${modulesAndAll[*]} " =~ ${optValue[module]} ]]; then
    usageExit "Invalid module: ${optValue[module]}"
  fi
elif [[ " ${repoCommands[*]} " =~ ${optValue[command]} ]]; then
  providedArgs=()
  wantedArgs=${repoCommandArgs[${repoCommandValues[${optValue[command]}]}]}

  for arg in "${optValue[module]}" "${@:3}"; do
    if [ -n "$arg" ]; then
      providedArgs+=("$arg")
    fi
  done

  if [ ${#providedArgs[@]} != "$wantedArgs" ]; then
    usageExit "Command ${optValue[command]} wants $wantedArgs arguments; received ${#providedArgs[@]} (${providedArgs[*]})"
  fi
fi
### script ###

case ${optValue[command]} in
"${moduleCommands[lint]}")
  runCommandInModule lint "${optValue[module]}"
  ;;
"${moduleCommands[lintfix]}")
  runCommandInModule lintFix "${optValue[module]}"
  ;;
"${moduleCommands[format]}")
  runCommandInModule format "${optValue[module]}"
  ;;
"${moduleCommands[test]}")
  runCommandInModule unitTests "${optValue[module]}"
  ;;
"${moduleCommands[build]}")
  runCommandInModule build "${optValue[module]}"
  ;;
"${repoCommands[ci]}")
  ci
  ;;
"${repoCommands[formatConfigs]}")
  formatConfigs
  ;;
"${repoCommands[spellcheck]}")
  spellcheckDocs
  ;;
"${repoCommands[setup]}")
  setup
  ;;
"${repoCommands[testSh]}")
  testScripts
  ;;
"${repoCommands[tag]}")
  tag "${@:2}"
  ;;
"${repoCommands[mocks]}")
  generateMocks
  ;;
"${repoCommands[coverage]}")
  coverage
  ;;
"${repoCommands[goWork]}")
  goWork
  ;;
"${repoCommands[newMod]}")
  newMod "${@:2}"
  ;;
esac
