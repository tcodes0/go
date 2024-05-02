#! /usr/bin/env bash

set -euo pipefail

# we use find to find folders directly under ./src/that have at least 1 *.go file
read -r -a packages <<<"$(find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed 's/^src\///' | tr '\n' '|')"
commands=(
  lint    # 0
  lintfix # 1
  format  # 2
  test    # 3
  build   # 4
)

usageExit() {
  echo "Usage: $0 [$(
    IFS=\|
    echo "${commands[*]}"
  )] [${packages[*]}]"
  exit 1
}

if [ $# -lt 2 ]; then
  echo "Invalid number of arguments $# ($*)"
  usageExit
fi

commandArg=$1
packageArg=$2

if ! [[ " ${commands[*]} " =~ $commandArg ]]; then
  echo "Invalid command: $commandArg"
  usageExit
fi

if ! [[ " ${packages[*]} " =~ $packageArg ]]; then
  echo "Invalid package: $packageArg"
  usageExit
fi

lintFlags=(--timeout 10s --print-issued-lines=false)
prefix=src/
prefixedPkgArg=$prefix$packageArg

runLint() {
  golangci-lint run "${lintFlags[@]}" "$prefixedPkgArg"
}

runLintFix() {
  ./src/sh/lint-fix.sh "$prefixedPkgArg"
}

runFormat() {
  gofumpt -l -w "$prefixedPkgArg"
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
esac
