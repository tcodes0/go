#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
trap 'err $LINENO' ERR

##########################
### vars and functions ###
##########################

usage() {
  command cat <<-EOF
Usage:
Run tests and output coverage files

$0 <module> (required)
EOF
}

test_directory() {
  # extract test package name from path
  # shellcheck disable=SC2155
  local test_pkg=$(basename "$MOD_PATH")_test reg_exp_prefix_cmd="^cmd/"
  local test_dir="./$MOD_PATH/$test_pkg"

  if [[ "$MOD_PATH" =~ $reg_exp_prefix_cmd ]]; then
    # cmds don't follow _test subpackage convention
    test_dir="./$MOD_PATH"
  fi

  if ! [ -d "$test_dir" ]; then
    # some packages have no tests
    exit 0
  fi

  echo -n "$test_dir"
}

run_tests() {
  # shellcheck disable=SC2155
  local test_dir=$(test_directory) test_output_json=$(mktemp /tmp/go-test-json-XXXXXX) reg_exp_prefix_go="^go:"

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

  # ignore failure to continue script
  go test "${flags[@]}" "$test_dir" >"$test_output_json" || true

  # delete lines not parseable as json output from 'go test'
  $SED --in-place --regexp-extended -e "/$reg_exp_prefix_go/d" "$test_output_json"

  gotestfmt -input "$test_output_json"

  echo "test_output_json=$test_output_json"
  echo "test_output_json=$test_output_json" >>"$GITHUB_OUTPUT"
}

coverage() {
  if [ ! -f "$COVERAGE_FILE" ]; then
    fatal $LINENO "$COVERAGE_FILE not found, did you run tests with -coverprofile=?"
  fi

  cover -html="$COVERAGE_FILE" -o coverage.html

  local opener=xdg-open

  if macos; then
    opener=open
  fi

  msgln view coverage html: \'$opener "$PWD/coverage.html"\'
}

##############
### script ###
##############

if requested_help "$*"; then
  usage
  exit 1
fi

run_tests
coverage
