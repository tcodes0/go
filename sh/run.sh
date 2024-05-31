#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

declare -rA opts=(
  ["all"]="all"
)

read -rd "$CHAR_CARRIG_RET" -a packages < <(
  printf '%b ' "${opts[all]}"

  regExpDotSlashPrefix="^\.\/"
  # find folders directly under . that have at least 1 *.go file; prettify output
  packages | sed -e "s/$regExpDotSlashPrefix//" | tr '\n' ' '

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
  msg "Usage: $0 <repo command>"
  msg "Usage: $0 <package command> <package>"
  msg "repo commands:\n\t$(joinBy '\n\t' "${repoCommands[@]}")"
  msg "package commands:\n\t$(joinBy '\n\t' "${packageCommands[@]}")"
  msg "packages:\n\t$(joinBy '\n\t' "${packages[@]}")"

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
  PKG_PATH="$1" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./sh/workflows/package-pr/test-pretty.sh
}

build() {
  PKG_PATH="$1" \
    ./sh/workflows/package-pr/build-go.sh && echo ok
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
  local package=$2

  $command "$package" || true

  if [ -d "$PWD/$package/${package}_test" ]; then
    $command "$package/${package}_test" || true
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
  ./sh/mocks.sh
}

coverage() {
  ./sh/coverage.sh
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "One or more arguments required"
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
  if [ "${optValue[package]}" ] && [ "${optValue[command]}" != "${repoCommands["tag"]}" ]; then
    usageExit "Command ${optValue[command]} takes no arguments; received ${*:2}"
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
