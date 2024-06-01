#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

# fail if any dependencies are missing
flags+=(-mod=readonly)
# verbose
flags+=(-v)

# building tests without regular .go files will fail
if ! [[ "$MOD_PATH" =~ test$ ]]; then
  go build "${flags[@]}" "./$MOD_PATH"
fi
