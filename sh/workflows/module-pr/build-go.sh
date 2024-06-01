#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

# fail if any dependencies are missing
flags+=(-mod=readonly)
# verbose
flags+=(-v)

go build "${flags[@]}" "./$MOD_PATH"
