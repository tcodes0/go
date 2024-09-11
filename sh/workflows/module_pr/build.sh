#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar nullglob
trap 'err $LINENO' ERR

##########################
### vars and functions ###
##########################

build_dir=".build"

usage() {
  command cat <<-EOF
Usage:
$(basename "$0") <module>             build the module on $build_dir    (required)
$(basename "$0") <module> -install    build and install to \$GOPATH
EOF
}

main() {
  local package_path="$1" command=build flags=()

  # building tests will fail
  if [[ "$package_path" =~ test$ ]]; then
    exit 0
  fi

  # cd to $build_dir/
  flags+=(-C "$build_dir")
  # fail if any dependencies are missing
  flags+=(-mod=readonly)
  # verbose
  flags+=(-v)
  # detect race conditions
  flags+=(-race)
  # path to build
  flags+=("../$package_path")

  if [[ "$*" =~ -install ]]; then
    msg "installing build"
    command=install
  fi

  mkdir -p "$build_dir"
  go $command "${flags[@]}"
}

##############
### script ###
##############

if requested_help "$*"; then
  usage
  exit 1
fi

main "$@"
