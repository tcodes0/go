#! /usr/bin/env bash

####################################
## this script is sourced by path ##
## from other scripts, careful    ##
## if moving or renaming it       ##
####################################

set -euo pipefail
shopt -s globstar

export PASS_GREEN="\e[7;38;05;029m PASS \e[0m"
export FAIL_RED="\e[2;7;38;05;197;47m FAIL \e[0m"
export COLOR_DIM="\e[2m"
export COLOR_END="\e[0m"
export CHAR_CARRIG_RET
CHAR_CARRIG_RET=$(printf "%b" "\r")
export COVERAGE_FILE="coverage.out"

# example: msg hello world
msg() {
  echo -e "> $*"
}

# example: msgExit could not find the file
msgExit() {
  msg "$*"
  return 1
}

# example: requireGitClean please commit changes to avoid losing work
requireGitClean() {
  message="${*:-There are uncommitted changes, please commit or stash}"

  if [ -n "$(git diff --exit-code)" ]; then
    msgExit "$message"
  fi
}

# output example: "23". Lines are terminal Y axis
currentTerminalLine() {
  # https://github.com/dylanaraps/pure-bash-bible#get-the-current-cursor-position
  IFS='[;' read -p $'\e[6n' -d R -rs _ currentLine _ _
  printf "%s" "$currentLine"
}

# example: requireInternet Internet required to fetch dependencies
requireInternet() {
  declare -A pingPals=(["cloudflare"]="1.1.1.1")
  message="${*:-Internet required}"

  if ! ping -c 1 "${pingPals[cloudflare]}" &>/dev/null; then
    msgExit "$message"
  fi
}

# run a test case and print the result
testCase() {
  local description=$1
  local input=$2
  local expected=$3
  local result

  # let the command expand
  # shellcheck disable=SC2086
  if ! result=$($TESTEE $input); then
    printf "%b\n" "$FAIL_RED $description"
    printf "%b\n" "non zero exit"
    exit 1
  fi

  if [ "$result" != "$expected" ]; then
    printf "%b\n" "$FAIL_RED $description"
    printf "%b\n" "expectation not met:"
    printf "%b\n" "< expected"
    diff <(printf %b "$expected") <(printf %b "$result")
    exit 1
  fi

  printf "%b\n" "$PASS_GREEN $description"
}

# wait for all processes to finish
# example: wait 123 345 5665 3234
wait() {
  while true; do
    done=$#

    for pid in "${@}"; do
      if ! ps -p "$pid" >/dev/null; then
        done=$((done - 1))
      fi
    done

    if [ "$done" = 0 ]; then
      break
    fi
  done
}

# example: requireGitBranch main
requireGitBranch() {
  branch="${1}"
  current=$(git branch --show-current)

  if [ "$branch" != "$current" ]; then
    msgExit "Please checkout $branch; on $current"
  fi
}

# find all packages in the project by looking for go files
packages() {
  find . -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort --stable | uniq
}

# example: joinBy , a b c. output: a, b, c
joinBy() {
  local delim=${1:-} first=${2:-}

  if shift 2; then
    printf %s "$first" "${@/#/$delim}"
  fi
}
