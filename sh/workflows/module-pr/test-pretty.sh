#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

# mod path may have a -h
if requestedHelp "$MOD_PATH"; then
  msgln "Usage:"
  msgln "$0 <module>"
  msgln "$0 <module> -cover"
  exit 1
fi

# extract test package name from path
testPkg=$(basename "$MOD_PATH")_test
testDir="./$MOD_PATH/$testPkg"
regExpPrefixCmd="^cmd/"

if [[ "$MOD_PATH" =~ $regExpPrefixCmd ]]; then
  # cmds have just a main package
  testDir="./$MOD_PATH"
fi

if ! [ -d "$testDir" ]; then
  # some packages have no tests
  exit 0
fi

# fail if any dependencies are missing
flags+=(-mod=readonly)
# output test results in json format for processing
flags+=(-json)
# detect race conditions
flags+=(-race)
# go vet linter is handled by lint step
flags+=(-vet=off)
# output coverage profile to file
flags+=(-coverprofile="$LIB_COVERAGE_FILE")
# package to scan coverage, necessary for blackbox testing
flags+=(-coverpkg="./$MOD_PATH")

if [ "$CACHE" == "false" ]; then
  # disable passed test caching
  flags+=(-count=1)
fi

testOutputJson=$(mktemp /tmp/go-test-json-XXXXXX)

# tee a copy of output for further processing
go test "${flags[@]}" "$testDir" 2>&1 | tee "$testOutputJson" | gotestfmt

# delete lines not parseable as json output from 'go test'
regExpPrefixGo="^go:"
_sed --in-place --regexp-extended -e "/$regExpPrefixGo/d" "$testOutputJson"

echo "testOutputJson=$testOutputJson"
echo "testOutputJson=$testOutputJson" >>"$GITHUB_OUTPUT"

if ! [ "${DISPLAY_COVERAGE:-}" ]; then
  exit
fi

if [ ! -f "$LIB_COVERAGE_FILE" ]; then
  msgln "$LIB_COVERAGE_FILE not found"
  exit 1
fi

cover -html="$LIB_COVERAGE_FILE" -o coverage.html

opener=xdg-open

if macos; then
  opener=open
fi

if $opener "$PWD/coverage.html" >/dev/null 2>&1; then
  msgln see your browser
else
  msgln "open $PWD/coverage.html.out in your browser"
fi
