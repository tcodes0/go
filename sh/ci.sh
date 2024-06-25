#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

start=$(date +%s)
failedJobs=""
passingJobs=0
ciPid=""

usageExit() {
  msgln "Usage: $0 "
  msgln "Usage: $0 push"
  exit 1
}

pushFailedJob() {
  local job="$1"

  if [[ "${failedJobs[*]}" != *${job}* ]]; then
    failedJobs+=" $job"
  fi
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
      passingJobs=$((passingJobs + 1))
    else
      printf "%b %b\n" "$LIB_COLOR_FAIL" "$job"
      pushFailedJob "$job"
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
  local failed log="$1" minDurationSeconds=5
  msgln

  if [ $(($(date +%s) - start)) -le $minDurationSeconds ]; then
    tac "$log" | head
    failed=true
  elif [ "${#failedJobs}" != 0 ]; then
    msgln ${#failedJobs} jobs failed \($passingJobs OK\)
    msgln see logs:

    for failed in "${failedJobs[@]}"; do
      msgln \'grep --color=always -Ee "$failedJobs" "$log"\'
    done

    failed=true
  elif [ "$passingJobs" == 0 ]; then
    grep --color=always -Eie "error" "$log" || true
    msgln "error: no jobs succeeded"
    failed=true
    # look for errors at the end of log, fail if found
  elif tac "$log" | head | grep --color=always -Eie error; then
    failed=true
  fi

  msgln
  msgln full logs:\\t\\t"$log"

  msgln
  msgln took $(($(date +%s) - start))s

  if [ "$failed" ]; then
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
