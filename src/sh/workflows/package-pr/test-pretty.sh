#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

# fail if any dependencies are missing
flags+=(-mod=readonly)
# output test results in json format for gotestfmt
flags+=(-json)
# detect race conditions
flags+=(-race)
# go vet linter is handled by lint step
flags+=(-vet=off)

if [ "$CACHE" == "false" ]; then
  # disable passed test caching
  flags+=(-count=1)
fi

testOutputJson=$(mktemp)

go test "${flags[@]}" "./$PKG/test" 2>&1 | tee "$testOutputJson" | gotestfmt

# a copy of test output is saved to a file for further processing in next workflow steps
# delete lines not parseable as json output from 'go test'
regExpPrefixGo="^go:"
sed -i "/$regExpPrefixGo/d" "$testOutputJson"

echo "testOutputJson=$testOutputJson"
echo "testOutputJson=$testOutputJson" >>"$GITHUB_OUTPUT"
