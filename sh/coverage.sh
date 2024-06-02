#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

### validation, input handling ###

### script ###

if [ ! -f "$COVERAGE_FILE" ]; then
  msgln "$COVERAGE_FILE not found, did you run tests?"
  exit 1
fi

cover -html="$COVERAGE_FILE" -o coverage.html.out

if xdg-open "$PWD/coverage.html.out" >/dev/null 2>&1 || open "$PWD/coverage.html.out" >/dev/null 2>&1; then
  msgln see your browser
  exit 0
fi

msgln "Open $PWD/coverage.html.out in your browser"
