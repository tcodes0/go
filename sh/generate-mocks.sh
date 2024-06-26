#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

mockeryGeneratedRegex="generated by mockery v[0-9]+\.[0-9]+\.[0-9]+"

saveMockeryVersion() {
  version=$(mockery --version 2>/dev/null)

  printf %s\\n "$version" >.mockery-version
}

stripVersion() {
  _sed --in-place --regexp-extended -e "s/$mockeryGeneratedRegex/generated by mockery/" ./**/mock*.go
}

renameExpect() {
  _sed --in-place --regexp-extended -e "s/EXPECT/Expect/" ./**/mock*.go
}

### validation, input handling ###

### script ###

if requestedHelp "$*"; then
  msgln "generate test mocks using mockery"
  exit 1
fi

mockery
saveMockeryVersion
stripVersion
renameExpect
