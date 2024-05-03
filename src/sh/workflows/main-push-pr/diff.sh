#! /usr/bin/env bash

set -euo pipefail

diff=$(git diff .)
if [ -n "$diff" ]; then
  echo "$diff"
  echo "update files and commit changes"
  exit 1
fi
