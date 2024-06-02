#! /usr/bin/env bash

########################################################
## this script is sourced by path from other scripts, ##
## careful if moving or renaming it                   ##
########################################################

set -euo pipefail
shopt -s globstar

export COLOR_PASS="\e[7;38;05;242m PASS \e[0m" COLOR_FAIL="\e[2;7;38;05;197;47m FAIL \e[0m" FORMAT_DIM="\e[2m"
export VISUAL_END="\e[0m" COVERAGE_FILE="coverage.out" ROOT_MODULE="github.com/tcodes0/go"

export CHAR_CARRIAGE_RET
CHAR_CARRIAGE_RET=$(printf "%b" "\r")

# example: msgln hello world
msgln() {
  msg "$*\\n"
}

# example: msg hello world
msg() {
  echo -ne "> $*"
}

# example: msgExit could not find the file
msgExit() {
  msgln "$*"
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
  local description=$1 input=$2 expected=$3 result

  # let the command expand
  # shellcheck disable=SC2086
  if ! result=$($TESTEE $input); then
    printf "%b\n" "$COLOR_FAIL $description"
    printf "%b\n" "non zero exit"
    exit 1
  fi

  if [ "$result" != "$expected" ]; then
    printf "%b\n" "$COLOR_FAIL $description"
    printf "%b\n" "expectation not met:"
    printf "%b\n" "< expected"
    diff <(printf %b "$expected") <(printf %b "$result")
    exit 1
  fi

  printf "%b\n" "$COLOR_PASS $description"
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
findModules() {
  find . -mindepth 2 -maxdepth 2 -type f -name '*.go' -exec dirname {} \; | sort --stable | uniq
}

# example: joinBy , a b c. output: a, b, c
joinBy() {
  local delim=${1:-} first=${2:-}

  if shift 2; then
    printf %s "$first" "${@/#/$delim}"
  fi
}

# example: requestedHelp "$*"
requestedHelp() {
  [[ "$*" =~ -h|--help|help ]] && return
  return 1
}
