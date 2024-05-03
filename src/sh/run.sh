#! /usr/bin/env bash

set -euo pipefail

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
commands=(
  lint          # 0
  lintfix       # 1
  format        # 2
  test          # 3
  build         # 4
  ci            # 5
  formatConfigs # 6
)

usage() {
  commandInfo=$(
    IFS=\|
    printf "%s" "${commands[*]}"
  )
  msg "Usage: $0 [$commandInfo] [${packages[*]}]"
}

if [ $# -lt 2 ]; then
  msgExit "Invalid number of arguments $# ($*)" "\n$(usage)"
fi

commandArg=$1
packageArg=$2

if ! [[ " ${commands[*]} " =~ $commandArg ]]; then
  msgExit "Invalid command: $commandArg"
fi

if ! [[ " ${packages[*]} " =~ $packageArg ]]; then
  msgExit "Invalid package: $packageArg"
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
"${commands[0]}")
  runLint "$prefixedPkgArg"
  ;;
"${commands[1]}")
  runLintFix "$prefixedPkgArg"
  ;;
"${commands[2]}")
  runFormat "$prefixedPkgArg"
  ;;
"${commands[3]}")
  runTest "$prefixedPkgArg"
  ;;
"${commands[4]}")
  runBuild "$prefixedPkgArg"
  ;;
"${commands[5]}")
  runCi
  ;;
"${commands[6]}")
  runFormatConfigs
  ;;
esac
