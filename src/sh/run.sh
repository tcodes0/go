#! /usr/bin/env bash

set -e -o pipefail

# we use find to find folders directly under ./src/that have at least 1 *.go file
read -r -a packages <<<"$(find src -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort | uniq | sed 's/^src\///' | tr '\n' '|')"
commands=(
  lint   # 0
  format # 1
  test   # 2
)
prefix=src/
commandArg=$1
packageArg=$2

usageExit() {
  echo "Usage: $0 [$(
    IFS=\|
    echo "${commands[*]}"
  )] [${packages[*]}]"
  exit 1
}

runLint() {
  echo runLint "$prefix$packageArg"
}

runFormat() {
  echo runFormat "$prefix$packageArg"
}

runTest() {
  echo runTest "$prefix$packageArg"
}

if [ $# -lt 2 ]; then
  echo "Invalid number of arguments $# ($*)"
  usageExit
fi

if ! [[ " ${commands[*]} " =~ $commandArg ]]; then
  echo "Invalid command: $commandArg"
  usageExit
fi

if ! [[ " ${packages[*]} " =~ $packageArg ]]; then
  echo "Invalid package: $packageArg"
  usageExit
fi

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
esac
