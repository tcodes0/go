#! /usr/bin/env bash

### options and imports ###

set -uo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

# run a test case and print the result
run() {
  local description=$1
  local input=$2
  local expected=$3
  local result

  # shellcheck disable=SC2086
  if ! result=$($TESTEE $input); then
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

# wait for all process to finish
# example: wait 123 345 5665 3234
wait() {
  while true; do
    doneCount=$#

    for pid in "${@}"; do
      if ! ps -p "$pid" >/dev/null; then
        doneCount=$((doneCount - 1))
      fi
    done

    if [ "$doneCount" = 0 ]; then
      break
    fi
  done
}

export GIT_TAG="v1.2.3"
export GIT_LOG="deadfaceb0 some commit message"
export EXEC_GIT=./src/sh/test/mocks/git.sh
export TESTEE=./src/sh/tag.sh
testsRunning=()
start=$(date +%s)

msg "$(basename $TESTEE)" test

run "increments major version" "major" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0
EOF
)" &
testsRunning+=($!)

run "increments minor version" "minor" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0
EOF
)" &
testsRunning+=($!)

run "increments patch version" "bump" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4
EOF
)" &
testsRunning+=($!)

GIT_TAG="v1.2.3-pre22"
run "increments patch pre release version" "bump" "$(
  cat <<EOF
current	v1.2.3-pre22
next	v1.2.3-pre23
EOF
)" &
testsRunning+=($!)

GIT_TAG="v1.2.3"
run "major pre release" "major -p" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0-pre1
EOF
)" &
testsRunning+=($!)

run "minor pre release" "minor -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0-pre1
EOF
)" &
testsRunning+=($!)

run "patch pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4-pre1
EOF
)" &
testsRunning+=($!)

GIT_TAG="v1.2.3-pre1"
run "patch already pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3-pre1
next	v1.2.3-pre2
EOF
)" &
testsRunning+=($!)

wait "${testsRunning[@]}"
msg took $(($(date +%s) - start))s
