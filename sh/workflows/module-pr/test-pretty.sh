#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

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
# flags+=(-race)
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

# tee a copy of output for further processing
go test "${flags[@]}" "$testDir" 2>&1 | tee "$testOutputJson" | gotestfmt

# delete lines not parseable as json output from 'go test'
regExpPrefixGo="^go:"
_sed --in-place -e "/$regExpPrefixGo/d" "$testOutputJson"

echo "testOutputJson=$testOutputJson"
echo "testOutputJson=$testOutputJson" >>"$GITHUB_OUTPUT"
