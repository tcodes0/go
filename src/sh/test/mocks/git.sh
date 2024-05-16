#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

case "$1" in
tag)
  if [ "$2" == --list ]; then
    printf %b "$MOCK_TAG"
  else
    true
  fi
  ;;
log)
  printf %b "$MOCK_LOG"
  ;;
show)
  printf %b "$MOCK_SHOW"
  ;;
*)
  printf %b "Command not mocked: $1"
  exit 1
  ;;
esac
