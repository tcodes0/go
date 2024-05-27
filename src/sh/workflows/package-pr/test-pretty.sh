#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

# extract package name from path
testPkg=$(basename "$PKG_PATH")_test
testDir="./$PKG_PATH/$testPkg"

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
flags+=(-coverpkg="./$PKG_PATH")

if [ "$CACHE" == "false" ]; then
  # disable passed test caching
  flags+=(-count=1)
fi

testOutputJson=$(mktemp /tmp/go-test-json-XXXXXX)

# tee a copy of output for further processing
go test "${flags[@]}" "$testDir" 2>&1 | tee "$testOutputJson" | gotestfmt

# delete lines not parseable as json output from 'go test'
regExpPrefixGo="^go:"
sed -Ei "/$regExpPrefixGo/d" "$testOutputJson"

echo "testOutputJson=$testOutputJson"
echo "testOutputJson=$testOutputJson" >>"$GITHUB_OUTPUT"
