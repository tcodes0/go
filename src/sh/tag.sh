#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a tags < <(
  git tag --list --sort=-refname | head
  printf %b "$CHAR_CARRIG_RET"
)
IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a log < <(
  git log --oneline --decorate | head
  printf %b "$CHAR_CARRIG_RET"
)

