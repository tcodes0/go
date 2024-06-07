#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

buildDir=".build"

# cd to $buildDir/
flags+=(-C "$buildDir")
# fail if any dependencies are missing
flags+=(-mod=readonly)
# verbose
flags+=(-v)
# detect race conditions
flags+=(-race)

if requestedHelp "$*"; then
  msgln "Usage: $0 <module-path>"
  exit 1
fi

command=build

if [ "${INSTALL:-}" ]; then
  command=install
fi

# building tests without regular .go files will fail
if ! [[ "$MOD_PATH" =~ test$ ]]; then
  mkdir -p "$buildDir"
  go $command "${flags[@]}" "../$MOD_PATH"
fi
