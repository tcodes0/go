#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar

### vars and functions ###

### validation, input handling ###

### script ###

if ! command -v commitlint >/dev/null; then
  npm install --global @commitlint/cli@"$VERSION"
fi

commitlint --config="$CONFIG_PATH" <<<"$(git log --format=%B -n 1 HEAD)"
