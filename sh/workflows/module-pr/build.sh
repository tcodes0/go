#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck source=../../lib.sh
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
  msgln "Inputs:"
  msgln "<module>\t build the module on $buildDir\t (required)"
  msgln "-install\t install build output"
  exit 1
fi

command=build

if [[ "$*" =~ -install ]]; then
  msg "installing build"
  command=install
fi

# building tests will fail
if [[ "$MOD_PATH" =~ test$ ]]; then
  exit 0
fi

# module has a directory for main.go, adjust build flags
# flags are relative to buildDir, MODPATH is not
if [ -d "$MOD_PATH/main" ]; then
  flags+=(-o "$(basename "$MOD_PATH")")
  flags+=("../$MOD_PATH/main")
else
  flags+=("../$MOD_PATH")
fi

mkdir -p "$buildDir"
go $command "${flags[@]}"
