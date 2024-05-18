#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

regExpSrcPrefix="^src\/"
# find folders directly under ./src that have at least 1 *.go file; massage the output a bit
packages="$(find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed -e "s/$regExpSrcPrefix//" | tr '\n' ' ')"
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

usageExit() {
  msg "$*\n"
  msg "Usage: $0 [${packageCommands[*]}] [$packages]"
  msg "Usage: $0 [${repoCommands[*]}]"

  exit 1
}

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

  if ! [[ " $packages " =~ $packageArg ]]; then
    usageExit "Invalid package: $packageArg"
  fi
fi

lintFlags=(--timeout 10s --print-issued-lines=false)
prefixedPkgArg=src/$packageArg
prettierFileGlob="**/*{.yml,.yaml,.json}"

runLint() {
  golangci-lint run "${lintFlags[@]}" "$prefixedPkgArg"
}

runLintFix() {
  requireGitClean
  ./src/sh/lint-fix.sh "$prefixedPkgArg"
}

runFormat() {
  requireGitClean
  gofumpt -l -w "$prefixedPkgArg"
  prettier --write "$prefixedPkgArg/$prettierFileGlob" 2>/dev/null || true
}

runFormatConfigs() {
  requireGitClean
  prettier --write "./$prettierFileGlob" 2>/dev/null || true
}

runTest() {
  PKG="$prefixedPkgArg" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./src/sh/workflows/package-pr/test-pretty.sh
}

runBuild() {
  PKG="$prefixedPkgArg" \
    ./src/sh/workflows/package-pr/build-go.sh && echo ok
}

runCi() {
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

case $commandArg in
"${packageCommands["lint"]}")
  runLint "$prefixedPkgArg"
  ;;
"${packageCommands["lintfix"]}")
  runLintFix "$prefixedPkgArg"
  ;;
"${packageCommands["format"]}")
  runFormat "$prefixedPkgArg"
  ;;
"${packageCommands["test"]}")
  runTest "$prefixedPkgArg"
  ;;
"${packageCommands["build"]}")
  runBuild "$prefixedPkgArg"
  ;;
"${repoCommands["ci"]}")
  runCi
  ;;
"${repoCommands["format"]}")
  runFormatConfigs
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
