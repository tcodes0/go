#! /usr/bin/env bash

set -euo pipefail

# fail if any dependencies are missing
flags+=(-mod=readonly)
# verbose
flags+=(-v)

go build "${flags[@]}" "./$PKG"
