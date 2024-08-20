#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck source=../../lib.sh
source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

### vars and functions ###

release_re="chore:[[:blank:]]+release"
changelog_h1_re="#[[:blank:]]+([[:alnum:]]+):[[:blank:]]+(v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+)[[:blank:]]+\*\(([[:digit:]]+-[[:digit:]]+-[[:digit:]]+)"

validate() {
  local main_head

  while read -r commit_line; do
    main_head="$commit_line"
    # unusual way to obtain the head of the main branch...
    # fix for SIGPIPE errors arrising from piping git log into head -1
    break
  done < <(git log --oneline --decorate)

  if ! [[ $main_head =~ $release_re ]]; then
    log "main head not a release commit: $main_head"
    exit 0
  fi
}

parse_changelog() {
  local tags=() newest_date module version date

  while read -r line; do
    if ! [[ "$line" =~ $changelog_h1_re ]]; then
      # want only release lines, h1 in md
      # one or more lines with modules to release
      continue
    fi

    module="${BASH_REMATCH[1]}"
    version="${BASH_REMATCH[2]}"
    date="${BASH_REMATCH[3]}"

    if [ ! "${newest_date:-}" ]; then
      # want to push tags for the newest release only
      # which will be in the top of the changelog
      newest_date=$date
    fi

    if [ "$date" != "$newest_date" ]; then
      # newest release is over
      break
    fi

    tags+=("$module/$version")
  done <"$CHANGELOG_FILE"

  printf %s "${tags[*]}"
}

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

push_tags() {
  for tag in "$@"; do
    msgln "$tag"
    git tag "$tag" HEAD
  done

  git push origin --tags
}

### script ###

validate
parsed_tags="$(parse_changelog)"
# shellcheck disable=SC2068 # intentional splitting
validate_tags ${parsed_tags[@]}
# shellcheck disable=SC2068 # intentional splitting
push_tags ${parsed_tags[@]}
