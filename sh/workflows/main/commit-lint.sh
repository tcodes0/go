#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar

### vars and functions ###

CONVENTIONAL_COMMITS_URL="See https://www.conventionalcommits.org/en/v1.0.0/"

lintCommits() {
  log="$1" problems=()

  while read -r commit; do
    out="$(commitlint --config="$CONFIG_PATH" <<<"$commit" || true)"

    if [ -n "$out" ]; then
      problems+=("$out")
    fi
  done <<<"$log"

  printf %s "${problems[*]}"
}

### validation, input handling ###

if [ -z "${BASE_REF:-}" ]; then
  BASE_REF=main
fi

### script ###

if ! command -v commitlint >/dev/null; then
  npm install --global @commitlint/cli@"$VERSION" >/dev/null
fi

if [ -n "${PR_TITLE:-}" ]; then
  echo "$PR_TITLE"

  if ! commitlint --config="$CONFIG_PATH" <<<"$PR_TITLE"; then
    echo "PR title must be a conventional commit, got: $PR_TITLE"
    echo "$CONVENTIONAL_COMMITS_URL"
    exit 1
  fi

  echo "PR title ok"
fi

revision=refs/remotes/origin/"$BASE_REF"..HEAD

echo git log "$revision"

log=$(git log --format=%s "$revision" --)

if [ -z "$log" ]; then
  echo "empty git log"
  exit
fi

echo "$log"

issues=$(lintCommits "$log")

if [ -n "$issues" ]; then
  totalCommits=$(wc -l <<<"$log")
  badCommits=$(grep -Eie input -c <<<"$issues")

  echo commits:
  echo "$log"
  echo
  echo linter\ output:
  echo
  echo "$issues"
  echo
  echo "Commit messages not formatted properly: $badCommits out of $totalCommits commits"
  echo "$CONVENTIONAL_COMMITS_URL"
  echo "To fix all, try 'git rebase -i $revision', change bad commits to 'reword', fix messages and 'git push --force'"
  exit 1
fi

echo "commits ok"
