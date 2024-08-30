#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar

if requested_help "$*"; then
  msgln "Inputs:"
  msgln "<module>\t run tests; output coverage files\t (required)"
  exit 1
fi

# extract test package name from path
testPkg=$(basename "$MOD_PATH")_test
testDir="./$MOD_PATH/$testPkg"
regExpPrefixCmd="^cmd/"

if [[ "$MOD_PATH" =~ $regExpPrefixCmd ]]; then
  # cmds don't follow _test subpackage convention
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
flags+=(-coverprofile="$COVERAGE_FILE")
# package to scan coverage, necessary for blackbox testing
flags+=(-coverpkg="./$MOD_PATH")

if [ "$CACHE" == "false" ]; then
  # disable passed test caching
  flags+=(-count=1)
fi

testOutputJson=$(mktemp /tmp/go-test-json-XXXXXX)

# ignore failure to continue script
go test "${flags[@]}" "$testDir" >"$testOutputJson" || true

# delete lines not parseable as json output from 'go test'
regExpPrefixGo="^go:"
$SED --in-place --regexp-extended -e "/$regExpPrefixGo/d" "$testOutputJson"

gotestfmt -input "$testOutputJson"

echo "testOutputJson=$testOutputJson"
echo "testOutputJson=$testOutputJson" >>"$GITHUB_OUTPUT"

if [ ! -f "$COVERAGE_FILE" ]; then
  fatal $LINENO "$COVERAGE_FILE not found, did you run tests with -coverprofile=?"
fi

cover -html="$COVERAGE_FILE" -o coverage.html

opener=xdg-open

if macos; then
  opener=open
fi

msgln view coverage html: \'$opener "$PWD/coverage.html"\'
