#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

start=$(date +%s)
firstFailedJob=""
hasSuccessfulJob=""
ciPid=""

usageExit() {
  msgln "Usage: $0 "
  msgln "Usage: $0 push"
  exit 1
}

# $1 logfile
printJobProgress() {
  local success="succeeded" failed="failed" ciLog="$1"
  local regExpJobStatus="\[([^]]+)\].*(succeeded|failed)"

  while read -r line; do
    if [[ ! "$line" =~ $regExpJobStatus ]]; then
      continue
    fi

    local job="${BASH_REMATCH[1]}" status="${BASH_REMATCH[2]}"

    if [ "$status" == "$success" ]; then
      printf "%b %b%s%b\n" "$LIB_COLOR_PASS" "$LIB_FORMAT_DIM" "$job" "$LIB_VISUAL_END"
      hasSuccessfulJob=true
    else
      printf "%b %b\n" "$LIB_COLOR_FAIL" "$job"
      if [ ! "$firstFailedJob" ]; then
        firstFailedJob="$job"
      fi
    fi
  done < <(grep -Eie "Job ($success|$failed)" "$ciLog" || true)
}

# script args
validateInput() {
  if [ $# -gt 1 ]; then
    msgln "Invalid arguments: $*"
    usageExit
  fi
}

# $1 "push" or empty for PR
prepareLogs() {
  local gitLocalBranch
  gitLocalBranch=$(git branch --show-current)
  local prJson="
{
  \"pull_request\": {
    \"title\": \"feat(ci): add PR title to act event\",
    \"head\": {
      \"ref\": \"$gitLocalBranch\"
    },
    \"base\": {
      \"ref\": \"main\"
    }
  },
  \"local\": true
}
"
  local pushJson="
{
  \"push\": {
    \"base_ref\": \"refs/heads/main\"
  },
  \"local\": true
}
"
  local eventJson="$prJson" eventJsonFile ciLog

  if [ "$1" == "push" ]; then
    eventJson="$pushJson"
  fi

  eventJsonFile=$(mktemp /tmp/ci-event-json-XXXXXX)
  ciLog=$(mktemp /tmp/ci-log-json-XXXXXX)

  printf "event json:" >"$ciLog"
  printf %s "$eventJson" >>"$ciLog"
  printf %s "$eventJson" >"$eventJsonFile"

  printf "%s %s" "$ciLog" "$eventJsonFile"
}

# $1 logfile
postCi() {
  local exitStatus=0 log="$1" minDurationSeconds=5
  msgln

  if [ $(($(date +%s) - start)) -le $minDurationSeconds ]; then
    tac "$log" | head
    exitStatus=1
  elif [ -n "$firstFailedJob" ]; then
    grep --color=always -Eie "$firstFailedJob" "$log" || true
    msgln "above: logs for '$firstFailedJob'"
    exitStatus=1
  elif [ -z "$hasSuccessfulJob" ]; then
    grep --color=always -Eie "error" "$log" || true
    msgln "error: no jobs succeeded"
    exitStatus=1
    # look for errors at end of log
  elif tac "$log" | head | grep --color=always -Eie error; then
    exitStatus=1
  fi

  msgln
  msgln full logs:\\t\\t"$log"

  msgln
  msgln took $(($(date +%s) - start))s

  if [ "$exitStatus" != 0 ]; then
    printf "%b" "$LIB_COLOR_FAIL"
  fi
}

### script ###

validateInput "$@"
read -rs logFile eventJsonFile <<<"$(prepareLogs "${1:-}")"

ciCommand="act"
ciCommandArgs=(-e "$eventJsonFile")
ciCommandArgs+=(-s GITHUB_TOKEN="$(gh auth token)")
ciCommandArgs+=(--container-architecture linux/amd64)

$ciCommand "${ciCommandArgs[@]}" >>"$logFile" 2>&1 || true &
ciPid=$!

printf "\e[H\e[2J" # move 1-1, clear whole screen
msgln "running ci..."

while ps -p "$ciPid" >/dev/null; do
  printf "\e[H" # move 1-1
  printJobProgress "$logFile"
  sleep 1
done

# catch status of last job that could have been missed by loop
printJobProgress "$logFile"
postCi "$logFile"
