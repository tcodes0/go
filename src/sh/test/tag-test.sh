#! /usr/bin/env bash

### options and imports ###

set -uo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

export TESTEE=./src/sh/tag.sh
export EXEC_GIT_READ=./src/sh/test/mocks/git.sh
export EXEC_GIT_WRITE=./src/sh/test/mocks/git.sh
testsRunning=()
start=$(date +%s)

### script ###

msg "$(basename $TESTEE)" test

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "increments major version" "major" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v2.0.0
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "increments minor version" "minor" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.3.0
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "increments patch version" "bump" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.4
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "major pre release" "major -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v2.0.0-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "minor pre release" "minor -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.3.0-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "patch pre release" "bump -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.4-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre22 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "increments pre release version" "bump" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.3-pre23
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre1 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "patch already a pre release" "bump -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.3-pre2
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_LOG="bada55c0ffe (HEAD -> main) hello world" \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "dry run: increments patch version" "bump -n" "$(
  cat <<EOF
git tag v1.2.4
deada55000 cactus (HEAD)
tagged with v1.2.4
EOF
)" &
testsRunning+=($!)

wait "${testsRunning[@]}"
msg took $(($(date +%s) - start))s
