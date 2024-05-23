#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

### validation, input handling ###

### script ###

if [ ! -f "$COVERAGE_FILE" ]; then
  msg "$COVERAGE_FILE not found, did you run tests?"
  exit 1
fi

cover -html="$COVERAGE_FILE" -o coverage.html.out

if command -v xdg-open &>/dev/null; then
  xdg-open "$PWD/coverage.html.out"
  msg see your browser
  exit 0
fi

if command -v open &>/dev/null; then
  open "$PWD/coverage.html.out"
  msg see your browser
  exit 0
fi

msg "Open $PWD/coverage.html.out in your browser"
