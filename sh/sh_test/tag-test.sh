#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -uo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

export TESTEE=./sh/tag.sh
export EXEC_GIT_READ=./sh/sh_test/mocks/git.sh
export EXEC_GIT_WRITE=./sh/sh_test/mocks/git.sh
testsRunning=()
start=$(date +%s)

### script ###

msgln "$(basename $TESTEE)" test

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "bump major version" "major" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "bump minor version" "minor" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "bump patch version" "bump" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "major pre release" "major -p" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "minor pre release" "minor -p" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "patch pre release" "bump -p" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre22 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3-pre22)\n" \
  testCase "bump pre release version" "bump" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3-pre22)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3-pre1 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3-pre1)\n" \
  testCase "patch already a pre release" "bump -p" "$(
  cat <<EOF
deada55000 cactus (HEAD -> tag: v1.2.3-pre1)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="deada55000 cactus (HEAD -> tag: v1.2.3)\n" \
  testCase "dry run: bump patch version" "bump -n" "$(
  cat <<EOF
dry run: git tag v1.2.4 HEAD
deada55000 cactus (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

MOCK_TAG=v1.2.3 \
  MOCK_SHOW="bada55c0ffe decaff please (HEAD -> tag: v1.2.3)\n" \
  testCase "bump patch version of commit" "bump -c bada55c0ffe" "$(
  cat <<EOF
bada55c0ffe decaff please (HEAD -> tag: v1.2.3)
EOF
)" &
testsRunning+=($!)

wait "${testsRunning[@]}"
msgln took $(($(date +%s) - start))s
