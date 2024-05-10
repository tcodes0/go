#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

# we use find to find folders directly under ./src that have at least 1 *.go file
regExpSrcPrefix="^src\/"
packages="$(find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed -e "s/$regExpSrcPrefix//" | tr '\n' '|')"
commandsWithArgs=(
  lint    # 0
  lintfix # 1
  format  # 2
  test    # 3
  build   # 4
)
commands=(
  ci             # 0
  formatConfigs  # 1
  spellcheckDocs # 2
  setup          # 3
)

usageExit() {
  commandArgsInfo=$(
    IFS=\|
    printf "%s" "${commandsWithArgs[*]}"
  )
  commandInfo=$(
    IFS=\|
    printf "%s" "${commands[*]}"
  )

  msg "$*\n"
  msg "Usage: $0 [$commandArgsInfo] [$packages]"
  msg "Usage: $0 [$commandInfo]"

  exit 1
}

if [ $# -lt 1 ]; then
  usageExit "Invalid number of arguments $# ($*)"
fi

commandArg=$1
packageArg=${2:-}

if ! [[ " ${commandsWithArgs[*]}${commands[*]} " =~ $commandArg ]]; then
  usageExit "Invalid command: $commandArg"
fi

if [[ " ${commandsWithArgs[*]} " =~ $commandArg ]]; then
  if [ -z "$packageArg" ]; then
    usageExit "Command $commandArg requires a package"
  fi

  if ! [[ " $packages " =~ $packageArg ]]; then
    usageExit "Invalid package: $packageArg"
  fi
fi

lintFlags=(--timeout 10s --print-issued-lines=false)
prefix=src/
prefixedPkgArg=$prefix$packageArg
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
  ./src/sh/ci/pull-request.sh
}

spellcheckDocs() {
  cspell "**/*.md" --gitignore
}

setup() {
  ./src/sh/setup.sh
}

case $commandArg in
"${commandsWithArgs[0]}")
  runLint "$prefixedPkgArg"
  ;;
"${commandsWithArgs[1]}")
  runLintFix "$prefixedPkgArg"
  ;;
"${commandsWithArgs[2]}")
  runFormat "$prefixedPkgArg"
  ;;
"${commandsWithArgs[3]}")
  runTest "$prefixedPkgArg"
  ;;
"${commandsWithArgs[4]}")
  runBuild "$prefixedPkgArg"
  ;;
"${commands[0]}")
  runCi
  ;;
"${commands[1]}")
  runFormatConfigs
  ;;
"${commands[2]}")
  spellcheckDocs
  ;;
"${commands[3]}")
  setup
  ;;
esac
