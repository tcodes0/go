#! /usr/bin/env bash

### options and imports ###

set -uo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

run() {
  local script=$1
  local description=$2
  local input=$3
  local expected=$4
  local result

  # shellcheck disable=SC2086
  if ! result=$($script $input); then
    printf "%b\n" "$FAIL_RED $description"
    printf "%b" "non zero exit"
    exit 1
  fi

  if [ "$result" != "$expected" ]; then
    printf "%b\n" "$FAIL_RED $description"
    printf "%b\n" "expectation not met:"
    printf "%b\n" "< expected"
    diff <(printf %b "$expected") <(printf %b "$result")
    exit 1
  fi

  printf "%b" "$PASS_GREEN $description"
}

export GIT_TAG="v1.2.3"
export GIT_LOG="deadfaceb0 some commit message"
export EXEC_GIT=./src/sh/test/mocks/git.sh

run ./src/sh/tag.sh "increments major version" "major" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0
EOF
)"
