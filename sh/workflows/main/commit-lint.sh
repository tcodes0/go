#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck source=lib.sh
source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

### vars and functions ###

CONVENTIONAL_COMMITS_URL="See https://www.conventionalcommits.org/en/v1.0.0/"

lintCommits() {
  local log="$1" problems="" out

  while read -r commit; do
    out="$(commitlint --config="$CONFIG_PATH" <<<"$commit" || true)"

    if [ "$out" ]; then
      problems+=("$out")
    fi
  done <<<"$log"

  printf %s "${problems[*]}"
}

validate() {
  if [ ! "${BASE_REF:-}" ]; then
    BASE_REF=main
  fi

  if ! command -v commitlint >/dev/null; then
    npm install --global @commitlint/cli@"$VERSION" >/dev/null
  fi
}

lintTitle() {
  if [ ! "${PR_TITLE:-}" ]; then
    return
  fi

  log "$PR_TITLE"

  if ! commitlint --config="$CONFIG_PATH" <<<"$PR_TITLE"; then
    msgln "PR title must be a conventional commit, got: $PR_TITLE"
    msgln "$CONVENTIONAL_COMMITS_URL"
    exit 1
  fi

  log "PR title ok"
}

lintLog() {
  local revision=refs/remotes/origin/"$BASE_REF"..HEAD

  log git log "$revision"

  log=$(git log --format=%s "$revision" --)

  if [ ! "$log" ]; then
    log "empty git log"
    return
  fi

  log "$log"

  issues=$(lintCommits "$log")

  if [ "$issues" ]; then
    totalCommits=$(wc -l <<<"$log")
    badCommits=$(grep -Eie input -c <<<"$issues")

    msgln commits:
    msgln "$log"
    msgln
    msgln linter\ output:
    msgln
    msgln "$issues"
    msgln
    msgln "Commit messages not formatted properly: $badCommits out of $totalCommits commits"
    msgln "$CONVENTIONAL_COMMITS_URL"
    msgln "To fix all, try 'git rebase -i $revision', change bad commits to 'reword', fix messages and 'git push --force'"

    return 1
  fi
}

### script ###

validate
lintTitle
lintLog

msgln "commits ok"
