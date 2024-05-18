#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

allPackages="all"
read -rd "$CHAR_CARRIG_RET" -a packages < <(
  printf %b "$allPackages "

  regExpSrcPrefix="^src\/"
  # find folders directly under ./src that have at least 1 *.go file; prettify output
  find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed -e "s/$regExpSrcPrefix//" | tr '\n' ' '

  printf %b "$CHAR_CARRIG_RET"
)

declare -rA packageCommands=(
  ["lint"]="lint"
  ["lintfix"]="lint-fix"
  ["format"]="format"
  ["test"]="test"
  ["build"]="build"
)

declare -rA repoCommands=(
  ["ci"]="ci"
  ["format"]="format-configs"
  ["spellcheck"]="spellcheck-docs"
  ["setup"]="setup"
  ["testSh"]="test-scripts"
  ["tag"]="tag"
  ["mocks"]="generate-mocks"
)

declare -A optValue=(
  # defaults
  ["all"]=""
)

usageExit() {
  msg "$*\n"
  msg "Usage: $0 [${packageCommands[*]}] [${packages[*]}]"
  msg "Usage: $0 [${repoCommands[*]}]"

  exit 1
}

lint() {
  local lintFlags=(--timeout 10s --print-issued-lines=false)
  golangci-lint run "${lintFlags[@]}" "$1"
}

lintFix() {
  if ! [ "${optValue[all]}" ]; then
    requireGitClean
  fi

  ./src/sh/lint-fix.sh "$1"
}

prettierFileGlob="**/*{.yml,.yaml,.json}"

format() {
  if ! [ "${optValue[all]}" ]; then
    requireGitClean
  fi

  gofumpt -l -w "$1"
  prettier --write "$1/$prettierFileGlob" 2>/dev/null || true
}

formatConfigs() {
  requireGitClean
  prettier --write "./$prettierFileGlob" 2>/dev/null || true
}

unitTests() {
  PKG="$1" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./src/sh/workflows/package-pr/test-pretty.sh
}

build() {
  PKG="$1" \
    ./src/sh/workflows/package-pr/build-go.sh && echo ok
}

ci() {
  requireGitClean
  requireInternet Internet required to pull docker images
  ./src/sh/ci/pull-request.sh
}

spellcheckDocs() {
  cspell "**/*.md" --gitignore
}

setup() {
  ./src/sh/setup.sh
}

testScripts() {
  find src/sh/test -iname "*-test.sh" -exec ./{} \;
}

tag() {
  requireGitBranch main
  ./src/sh/tag.sh "$@"
}

run() {
  local prefix="src/"
  if [ "${optValue[all]}" ]; then
    for pkg in "${packages[@]}"; do
      printf %b "\n"
      msg "$1 $pkg..."
      "$1" "$prefix$pkg" || true
    done
  else
    "$1" "$prefix$pkg"
  fi
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "Invalid number of arguments $# ($*)"
fi

commandArg=$1
packageArg=${2:-}

if ! [[ " ${packageCommands[*]}${repoCommands[*]} " =~ $commandArg ]]; then
  usageExit "Invalid command: $commandArg"
fi

if [[ " ${packageCommands[*]} " =~ $commandArg ]]; then
  if [ -z "$packageArg" ]; then
    usageExit "Command $commandArg requires a package"
  fi

  if ! [[ " ${packages[*]} " =~ $packageArg ]]; then
    usageExit "Invalid package: $packageArg"
  fi

  if [ "$packageArg" == "$allPackages" ]; then
    optValue["all"]=true
    packageArg=""
    packages=("${packages[@]:1}")
  fi
fi

if [[ " ${repoCommands[*]} " =~ $commandArg ]]; then
  if [ "$packageArg" ]; then
    usageExit "Command $commandArg takes no arguments"
  fi
fi

### script ###

case $commandArg in
"${packageCommands["lint"]}")
  run lint
  ;;
"${packageCommands["lintfix"]}")
  run lintFix
  ;;
"${packageCommands["format"]}")
  run format
  ;;
"${packageCommands["test"]}")
  run unitTests
  ;;
"${packageCommands["build"]}")
  run build
  ;;
"${repoCommands["ci"]}")
  ci
  ;;
"${repoCommands["format"]}")
  formatConfigs
  ;;
"${repoCommands["spellcheck"]}")
  spellcheckDocs
  ;;
"${repoCommands["setup"]}")
  setup
  ;;
"${repoCommands["testSh"]}")
  testScripts
  ;;
"${repoCommands["tag"]}")
  tag "${@:2}"
  ;;
"${repoCommands["mocks"]}")
  mockery
  ;;
esac
