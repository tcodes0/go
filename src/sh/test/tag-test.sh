#! /usr/bin/env bash

### options and imports ###

set -uo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

export TESTEE=./src/sh/tag.sh
export EXEC_GIT=./src/sh/test/mocks/git.sh
testsRunning=()
start=$(date +%s)

### script ###

msg "$(basename $TESTEE)" test

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "increments major version" "major" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "increments minor version" "minor" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "increments patch version" "bump" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "major pre release" "major -p" "$(
  cat <<EOF
current	v1.2.3
next	v2.0.0-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "minor pre release" "minor -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.3.0-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "patch pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3
next	v1.2.4-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre22 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "increments pre release version" "bump" "$(
  cat <<EOF
current	v1.2.3-pre22
next	v1.2.3-pre23
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre1 \
  MOCK_LOG="bada55c0ffe some commit message" \
  testCase "patch already a pre release" "bump -p" "$(
  cat <<EOF
current	v1.2.3-pre1
next	v1.2.3-pre2
EOF
)" &
testsRunning+=($!)

wait "${testsRunning[@]}"
msg took $(($(date +%s) - start))s
