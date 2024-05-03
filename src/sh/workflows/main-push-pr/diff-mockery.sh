#! /usr/bin/env bash

set -euo pipefail

mockery

diff=$(git diff .)
if [ -n "$diff" ]; then
  echo "$diff"
  echo "run 'mockery' and commit the changes"
  exit 1
fi
