#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

case "$1" in
tag)
  printf %b "$MOCK_TAG"
  ;;
log)
  printf %b "$MOCK_LOG"
  ;;
*)
  printf %b "Command not mocked: $1"
  exit 1
  ;;
esac
