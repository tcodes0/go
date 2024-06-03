#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar

### vars and functions ###

lintCommits() {
  local log="$1" problems=() out

  while read -r commit; do
    out="$(commitlint --config="$CONFIG_PATH" <<<"$commit" || true)"

    if [ -n "$out" ]; then
      problems+=("$out")
    fi
  done <<<"$log"

  printf %s "${problems[*]}"
}

### validation, input handling ###

### script ###

if ! command -v commitlint >/dev/null; then
  npm install --global @commitlint/cli@"$VERSION" >/dev/null
fi

if [ -z "$BASE_REF" ] || [ -z "$HEAD_REF" ]; then
  echo "BASE_REF or HEAD_REF are empty"
  exit 1
fi

if [ "$BASE_REF" == "$HEAD_REF" ]; then
  echo "No commits to lint"
  exit 0
fi

log=$(git log --format=%s origin/"$BASE_REF".."$HEAD_REF")
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
  echo "See https://www.conventionalcommits.org/en/v1.0.0/"
  echo "To fix all, try 'git rebase -i origin/main..HEAD', change bad commits to 'reword', fix messages and 'git push --force'"
  exit 1
fi
