#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.
#
# parses changelog and pushes tags the newest release

set -euo pipefail
shopt -s globstar
# shellcheck source=../../lib.sh
source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

##########################
### vars and functions ###
##########################

# Description: Exit the script if main head is doesn't match a release commit
# STDERR     : Logs
# Sideeffects: Might exit the script with success
# Example    : validate
validate() {
  local main_head release_re="chore:[[:blank:]]+release"

  while read -r commit_line; do
    main_head="$commit_line"
    # meant to only run once
    # fix for SIGPIPE errors arrising from piping git log into head -1
    break
  done < <(git log --oneline --decorate)

  if ! [[ $main_head =~ $release_re ]]; then
    log "main head not a release commit: $main_head"
    exit 0
  fi
}

# Description: Parses tags from changelog
# Globals    : CHANGELOG_FILE (github workflow or .env)
# STDOUT     : Parsed tags
# Returns    : Parsed tags
# Example    : tags="$(parse_changelog_tags)"
parse_changelog_tags() {
  local tags=() newest_date tag date
  # https://regex101.com/r/67rzeb/4
  local changelog_h1_re="#[[:blank:]]+(?:[a-zA-Z0-9 !%&*()-+]+:)?[[:blank:]]*([a-zA-Z0-9\/.]+)[[:blank:]]+\*\(([[:digit:]]+-[[:digit:]]+-[[:digit:]]+)"

  while read -r line; do
    if ! [[ "$line" =~ $changelog_h1_re ]]; then
      # want only release lines, h1 in md
      # one or more lines with tags
      continue
    fi

    tag="${BASH_REMATCH[1]}"
    date="${BASH_REMATCH[2]}"

    if [ ! "${newest_date:-}" ]; then
      # want to push tags for the newest release only
      # which will be in the top of the changelog
      newest_date=$date
    fi

    if [ "$date" != "$newest_date" ]; then
      # newest release is over
      break
    fi

    tags+=("$tag")
  done <"$CHANGELOG_FILE"

  printf %s "${tags[*]}"
}

# Description: Validates input as git tags
# Args       : Words separated by spaces
# STDOUT     : Help message
# STDERR     : Might log errors
# Sideeffects: Might exit with 1
# Example    : validate_tags foo/v1.1.1. bar/v1.2.3
validate_tags() {
  local tags=("$@")

  if [ ${#tags[@]} == 0 ]; then
    fatal "no tags found"
  fi

  for tag in "${tags[@]}"; do
    echo "$tag"
    if git rev-parse --verify "$tag" >/dev/null; then
      fatal "tag already exists: $tag"
    fi
  done
}

# Description: Creates and pushes git tags
# Args       : words separated by spaces
# STDOUT     : Prints each tag created, plus git output
# STDERR     : Git might output
# Sideeffects: Pushes git tags
# Example    : tag_and_push foo/v1.1.1. bar/v1.2.3
tag_and_push() {
  for tag in "$@"; do
    msgln "$tag"
    git tag "$tag" HEAD
  done

  git push origin --tags
}

##############
### script ###
##############

validate

if [ ! "${CHANGELOG_FILE-}" ]; then
  # shellcheck source=../../../.env
  source .env
fi

parsed_tags="$(parse_changelog_tags)"
# shellcheck disable=SC2068 # intentional splitting
validate_tags ${parsed_tags[@]}
# shellcheck disable=SC2068 # intentional splitting
tag_and_push ${parsed_tags[@]}
