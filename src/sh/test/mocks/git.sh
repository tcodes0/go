#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

if [ "$1" == "tag" ]; then
  printf %b "$GIT_TAG"
elif [ "$1" == "log" ]; then
  printf %b "$GIT_LOG"
fi
