#! /usr/bin/env bash

set -euo pipefail

# we use find to find folders directly under ./src/that have at least 1 *.go file
read -r -a packages <<<"$(find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed 's/^src\///' | tr '\n' '|')"
commands=(
  lint   # 0
  format # 1
  test   # 2
  build  # 3
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

prefix=src/
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

runLint() {
  golangci-lint run --timeout 10s --print-issued-lines=false "$prefix$packageArg"
}

runFormat() {
  gofumpt -l -w "$prefix$packageArg"
}

runTest() {
  PKG="$prefix$packageArg" \
    CACHE="true" \
    GITHUB_OUTPUT="/dev/null" \
    ./src/sh/workflows/package-pr/test-pretty.sh
}

runBuild() {
  PKG="$prefix$packageArg" \
    ./src/sh/workflows/package-pr/build-go.sh && echo ok
}

case $commandArg in
"${commands[0]}")
  runLint "$packageArg"
  ;;
"${commands[1]}")
  runFormat "$packageArg"
  ;;
"${commands[2]}")
  runTest "$packageArg"
  ;;
"${commands[3]}")
  runBuild "$packageArg"
  ;;
esac
