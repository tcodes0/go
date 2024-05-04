#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

import() {
  relativePath="go\/src\/sh\/shared-functions.sh"
  regex="(.*)\/go\/?.*" # \1 will capture base path
  functions=$(sed -E "s/$regex/\1\/$relativePath/g" <<<"$PWD")

  # shellcheck disable=SC1090
  source "$functions"
}

import

# we use find to find folders directly under ./src/that have at least 1 *.go file
read -r -a packages <<<"$(find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed 's/^src\///' | tr '\n' '|')"
commandsWithArgs=(
  lint    # 0
  lintfix # 1
  format  # 2
  test    # 3
  build   # 4
)
commands=(
  ci            # 0
  formatConfigs # 1
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
  msg "Usage: $0 [$commandArgsInfo] [${packages[*]}]"
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

  if ! [[ " ${packages[*]} " =~ $packageArg ]]; then
    usageExit "Invalid package: $packageArg"
  fi
fi

lintFlags=(--timeout 10s --print-issued-lines=false)
prefix=src/
prefixedPkgArg=$prefix$packageArg

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
  npx prettier --write "$prefixedPkgArg"
}

runFormatConfigs() {
  requireGitClean
  npx prettier --write ./**/*{.yml,.yaml,.json}
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
esac
