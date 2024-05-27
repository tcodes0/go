#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

declare -rA opts=(
  ["all"]="all"
)

read -rd "$CHAR_CARRIG_RET" -a packages < <(
  printf %b "${opts[all]} "

  regExpSrcPrefix="^src\/"
  # find folders directly under ./src that have at least 1 *.go file; prettify output
  find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort --stable | uniq | sed -e "s/$regExpSrcPrefix//" | tr '\n' ' '

  printf %b "$CHAR_CARRIG_RET"
)

declare -rA packageCommands=(
  ["build"]="build"
  ["format"]="format"
  ["lint"]="lint"
  ["lintfix"]="lint-fix"
  ["test"]="test"
)

declare -rA repoCommands=(
  ["ci"]="ci"
  ["coverage"]="coverage"
  ["formatConfigs"]="format-configs"
  ["mocks"]="generate-mocks"
  ["spellcheck"]="spellcheck-docs"
  ["setup"]="setup"
  ["tag"]="tag"
  ["testSh"]="test-scripts"
)

declare -A optValue=(
  # defaults
  ["all"]=""
  ["command"]=""
  ["package"]=""
)

usageExit() {
  msg "$*\n"
  msg "Usage: $0 [${packageCommands[*]}] [${packages[*]}]"
  msg "Usage: $0 [${repoCommands[*]}]"

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
  ./src/sh/lint-fix.sh "$path" &
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
  PKG_PATH="$1" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./src/sh/workflows/package-pr/test-pretty.sh
}

build() {
  PKG_PATH="$1" \
    ./src/sh/workflows/package-pr/build-go.sh && echo ok
}

ci() {
  requireGitClean
  requireInternet Internet required to pull docker images
  ./src/sh/ci.sh
}

spellcheckDocs() {
  cspell "**/*.md" --gitignore
}

setup() {
  ./src/sh/setup.sh
}

testScripts() {
  find src/sh/sh_test -iname "*-test.sh" -exec ./{} \;
}

tag() {
  requireGitBranch main
  ./src/sh/tag.sh "$@"
}

run() {
  local prefix="src"
  local command=$1
  local package=$2

  $command "${prefix}/$package" || true

  if [ -d "$PWD/${prefix}/$package/${package}_test" ]; then
    $command "${prefix}/$package/${package}_test" || true
  fi
}

runPkgCommand() {
  local command=$1

  if ! [ "${optValue[all]}" ]; then
    run "$command" "${optValue["package"]}"
    return
  fi

  for pkg in "${packages[@]}"; do
    printf %b "\n"
    msg "$command $pkg..."
    run "$command" "$pkg"
  done
}

generateMocks() {
  ./src/sh/mocks.sh
}

coverage() {
  ./src/sh/coverage.sh
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "Invalid number of arguments $# ($*)"
fi

optValue["command"]=$1
optValue["package"]=${2:-}

if ! [[ " ${packageCommands[*]}${repoCommands[*]} " =~ ${optValue[command]} ]]; then
  usageExit "Invalid command: ${optValue[command]}"
fi

if [[ " ${packageCommands[*]} " =~ ${optValue[command]} ]]; then
  if [ -z "${optValue[package]}" ]; then
    usageExit "Command ${optValue[command]} requires a package"
  fi

  if ! [[ " ${packages[*]} " =~ ${optValue[package]} ]]; then
    usageExit "Invalid package: ${optValue[package]}"
  fi

  if [ "${optValue[package]}" == "${opts[all]}" ]; then
    optValue["all"]=true
    optValue[package]=""
    packages=("${packages[@]:1}")
  fi
elif [[ " ${repoCommands[*]} " =~ ${optValue[command]} ]]; then
  if [ "${optValue[package]}" ]; then
    usageExit "Command ${optValue[command]} takes no arguments"
  fi
fi

### script ###

case ${optValue[command]} in
"${packageCommands[lint]}")
  runPkgCommand lint
  ;;
"${packageCommands[lintfix]}")
  runPkgCommand lintFix
  ;;
"${packageCommands[format]}")
  runPkgCommand format
  ;;
"${packageCommands[test]}")
  runPkgCommand unitTests
  ;;
"${packageCommands[build]}")
  runPkgCommand build
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
esac
