#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

# cd to build/
flags+=(-C .build)
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

# building tests without regular .go files will fail
if ! [[ "$MOD_PATH" =~ test$ ]]; then
  go build "${flags[@]}" "../$MOD_PATH"
fi
