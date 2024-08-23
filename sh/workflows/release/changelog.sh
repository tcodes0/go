#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.
#
# wrapper around changelog tool.
# Args: 1=url 2=title 3=prefix

set -euo pipefail
shopt -s globstar
# shellcheck source=../../lib.sh
source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

##########################
### vars and functions ###
##########################

# Description: Validates user input
# Globals    : CHANGELOG_FILE (github workflow or .env)
# Returns    : 1 if fail
# Example    : validate
validate() {
  if [ ! -f "$CHANGELOG_FILE" ]; then
    err $LINENO "$CHANGELOG_FILE" not found
    return 1
  fi
}

# Description: Calls changelog tool and updates CHANGELOG_FILE
# Globals    : CHANGELOG_FILE (github workflow or .env)
# Args       : 1=url 2=title 3=prefix
# STDERR     : Might print errors
# Returns    : 1 if fail
# Sideeffects: Updates CHANGELOG_FILE
# Example    : update_changelog pizza
update_changelog() {
  local url=$1 title=${2:-} prefix=${3:-} changes changelog flags=()

  changelog=$(cat "$CHANGELOG_FILE")
  flags+=(-title "$title")
  flags+=(-tagprefix "$prefix")
  flags+=(-url "$url")

  changes=$(changelog "${flags[@]}")
  if [ ! "$changes" ]; then
    err $LINENO "empty changes"
    return 1
  fi

  printf %s "$changes" >"$CHANGELOG_FILE"
  printf %s "$changelog" >>"$CHANGELOG_FILE"
}

##############
### script ###
##############

if [ ! "${CHANGELOG_FILE-}" ]; then
  # shellcheck source=../../.env
  source .env
fi

validate
update_changelog "$@"
