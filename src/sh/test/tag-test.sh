#! /usr/bin/env bash

### options and imports ###

set -uo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

export GIT_TAG="v1.2.3"
export GIT_LOG="bada55c0ffe some commit message"
export EXEC_GIT=./src/sh/test/mocks/git.sh
export TESTEE=./src/sh/tag.sh
testsRunning=()
start=$(date +%s)

### script ###

msg "$(basename $TESTEE)" test

testCase "increments major version" "major" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0
EOF
)" &
testsRunning+=($!)

testCase "increments minor version" "minor" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0
EOF
)" &
testsRunning+=($!)

testCase "increments patch version" "bump" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4
EOF
)" &
testsRunning+=($!)

GIT_TAG="v1.2.3-pre22"
testCase "increments pre release version" "bump" "$(
  cat <<EOF
current	v1.2.3-pre22
next	v1.2.3-pre23
EOF
)" &
testsRunning+=($!)

GIT_TAG="v1.2.3"
testCase "major pre release" "major -p" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0-pre1
EOF
)" &
testsRunning+=($!)

testCase "minor pre release" "minor -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0-pre1
EOF
)" &
testsRunning+=($!)

testCase "patch pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4-pre1
EOF
)" &
testsRunning+=($!)

GIT_TAG="v1.2.3-pre1"
testCase "patch already a pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3-pre1
next	v1.2.3-pre2
EOF
)" &
testsRunning+=($!)

wait "${testsRunning[@]}"
msg took $(($(date +%s) - start))s
