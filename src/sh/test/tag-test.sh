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
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "bump major version" "major" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v2.0.0
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "bump minor version" "minor" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.3.0
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "bump patch version" "bump" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.4
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "major pre release" "major -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v2.0.0-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "minor pre release" "minor -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.3.0-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "patch pre release" "bump -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.4-pre1
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre22 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "bump pre release version" "bump" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.3-pre23
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre1 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "patch already a pre release" "bump -p" "$(
  cat <<EOF
deada55000 cactus (HEAD)
tagged with v1.2.3-pre2
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD)\n" \
  testCase "dry run: bump patch version" "bump -n" "$(
  cat <<EOF
dry run: git tag v1.2.4 HEAD
deada55000 cactus (HEAD)
tagged with v1.2.4
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="bada55c0ffe decaff please (HEAD)\n" \
  testCase "bump patch version of commit" "bump -c bada55c0ffe" "$(
  cat <<EOF
bada55c0ffe decaff please (HEAD)
tagged with v1.2.4
EOF
)" &
testsRunning+=($!)

wait "${testsRunning[@]}"
msg took $(($(date +%s) - start))s
