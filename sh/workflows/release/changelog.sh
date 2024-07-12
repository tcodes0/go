#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
# source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

### vars and functions ###

changelog_file=CHANGELOG.md

validate() {
  if [ ! -f "$changelog_file" ]; then
    echo "$changelog_file" not found
    return 1
  fi
}

updateChangelog() {
  local module=$1 version=$2 old_version=$3 changes changelog flags=()

  changelog=$(cat "$changelog_file")
  flags+=(-title "$module: v$version")
  flags+=(-tag "$module/v$version")

  if [ "$old_version" ]; then
    flags+=(-old-tag "$module/v$old_version")
  fi

  changes=$(go run ./cmd/changelog/main.go "${flags[@]}")
  if [ ! "$changes" ]; then
    echo "empty changes"
    return 1
  fi

  printf %s "$changes" >"$changelog_file"
  printf %s "$changelog" >>"$changelog_file"
}

### script ###

module=$1
version=$2
old_version=$3

validate
updateChangelog "$module" "$version" "$old_version"
