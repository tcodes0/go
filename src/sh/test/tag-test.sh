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

  printf "%b\n" "$PASS_GREEN $description"
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

run ./src/sh/tag.sh "increments minor version" "minor" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0
EOF
)"

run ./src/sh/tag.sh "increments patch version" "bump" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4
EOF
)"

GIT_TAG="v1.2.3-pre22"
run ./src/sh/tag.sh "increments patch pre release version" "bump" "$(
  cat <<EOF
current	v1.2.3-pre22
next	v1.2.3-pre23
EOF
)"

GIT_TAG="v1.2.3"
run ./src/sh/tag.sh "major pre release" "major -p" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0-pre1
EOF
)"

run ./src/sh/tag.sh "minor pre release" "minor -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0-pre1
EOF
)"

run ./src/sh/tag.sh "patch pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4-pre1
EOF
)"

GIT_TAG="v1.2.3-pre1"
run ./src/sh/tag.sh "patch already pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3-pre1
next	v1.2.3-pre2
EOF
)"
