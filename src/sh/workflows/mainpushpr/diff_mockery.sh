#! /usr/bin/env bash

set -e
mockery

DIFF=$(git diff .)
if [ -n "$DIFF" ]; then
  echo "$DIFF"
  echo "run 'mockery' and commit the changes"
  exit 1
fi
