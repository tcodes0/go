#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar

### vars and functions ###

### validation, input handling ###

### script ###

diff=$(git diff .)

if [ -n "$diff" ]; then
  echo "$diff"
  echo "update files and commit changes"
  exit 1
fi
