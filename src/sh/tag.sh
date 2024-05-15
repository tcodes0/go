#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

# example: tag 1 2 3 pre 4 outputs v1.2.3-pre4
tag() {
  if [ -n "${4:-}" ]; then
    printf 'v%s.%s.%s-pre%s' "$1" "$2" "$3" "$5"
  else
    printf 'v%s.%s.%s' "$1" "$2" "$3"
  fi
}

IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a tags < <(
  git tag --list --sort=-refname | head
  printf %b "$CHAR_CARRIG_RET"
)
IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a logs < <(
  git log --oneline --decorate | head
  printf %b "$CHAR_CARRIG_RET"
)

latestTag="${tags[0]}"
regExpSemVerPre="v?([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+)-?(pre)?([[:digit:]]*)?"

if [[ "$latestTag" =~ $regExpSemVerPre ]]; then
  major="${BASH_REMATCH[1]}"
  minor="${BASH_REMATCH[2]}"
  patch="${BASH_REMATCH[3]}"
  preRelease="${BASH_REMATCH[4]}"
  PreReleaseVer="${BASH_REMATCH[5]}"
fi
