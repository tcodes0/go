#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

# fail if any dependencies are missing
flags+=(-mod=readonly)
# verbose
flags+=(-v)

if requestedHelp "$*"; then
  msgln "Usage: $0 <module-path>"
  exit 1
fi

# building tests without regular .go files will fail
if ! [[ "$MOD_PATH" =~ test$ ]]; then
  go build "${flags[@]}" "./$MOD_PATH"
fi
