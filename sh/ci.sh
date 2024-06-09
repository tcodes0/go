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

usageExit() {
  msgln "Usage: $0 "
  msgln "Usage: $0 push"
  exit 1
}

firstFailedJob=""
hasSuccessfulJob=""
gitLocalBranch=$(git branch --show-current)
inputArg="${1:-}"
eventJson=""

printJobProgress() {
  successToken="succeeded"
  failedToken="failed"
  ciLog="$1"

  tput rc
  grepOut=$(grep -Eie "Job ($successToken|$failedToken)" "$ciLog" || true)
  regExpAfterSpace=" .*"
  linesPrinted=$(wc -l <<<"$grepOut" | _sed --regexp-extended -e "s/$regExpAfterSpace//")

  if [ "$linesPrinted" != 0 ]; then
    while read -r line; do
      if [ -z "$line" ]; then
        continue
      fi

      if [[ "$line" =~ \[([^]]+)\].*(succeeded|failed) ]]; then
        job="${BASH_REMATCH[1]}"
        status="${BASH_REMATCH[2]}"

        if [ "$status" == "$successToken" ]; then
          printf "%b %b%s%b\n" "$LIB_COLOR_PASS" "$LIB_FORMAT_DIM" "$job" "$LIB_VISUAL_END"
          hasSuccessfulJob=true
        else
          printf "%b %b\n" "$LIB_COLOR_FAIL" "$job"
          if [ -z "$firstFailedJob" ]; then
            firstFailedJob="$job"
          fi
        fi
      fi
    done <<<"$grepOut"
  fi
}

prJson="
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

pushJson="
{
  \"push\": {
    \"base_ref\": \"refs/heads/main\"
  },
  \"local\": true
}
"

### validation, input handling ###

if [ $# -gt 1 ]; then
  msgln "Invalid arguments: $*"
  usageExit
fi

### script ###

if [ "$inputArg" ] && [ "$inputArg" == "push" ]; then
  eventJson="$pushJson"
else
  eventJson="$prJson"
fi

eventJsonFile=$(mktemp /tmp/ci-event-json-XXXXXX)
ciLog=$(mktemp /tmp/ci-log-json-XXXXXX)

printf "event json:" >"$ciLog"
printf %s "$eventJson" >>"$ciLog"
printf %s "$eventJson" >"$eventJsonFile"

ciCommand="act"
ciCommandArgs=(-e "$eventJsonFile")
ciCommandArgs+=(-s GITHUB_TOKEN="$(gh auth token)")
ciCommandArgs+=(--container-architecture linux/amd64)

$ciCommand "${ciCommandArgs[@]}" >>"$ciLog" 2>&1 || true &
ciPid=$!

lastLine=$(tput lines)
expectedOutput=35

if [ "$(currentTerminalLine)" -gt "$((lastLine - expectedOutput))" ]; then
  clear -x
  msgln "running ci... (terminal scrolled up to make room for output)"
else
  msgln "running ci..."
fi

tput sc

iterations=0
while ps -p $ciPid >/dev/null; do
  printJobProgress "$ciLog"
  iterations=$((iterations + 1))
  sleep 1
done

# catch status of last job that could have been missed by loop
printJobProgress "$ciLog"

exitStatus=0
somethingWrong=5
if [ "$iterations" -le "$somethingWrong" ]; then
  printf \\n
  tac "$ciLog" | head
  exitStatus=1
elif [ -n "$firstFailedJob" ]; then
  printf \\n
  grep --color=always -Eie "$firstFailedJob" "$ciLog" || true
  msgln "above: logs for '$firstFailedJob'"
  exitStatus=1
elif [ -z "$hasSuccessfulJob" ]; then
  printf \\n
  grep --color=always -Eie "error" "$ciLog" || true
  msgln "error: no jobs succeeded"
  exitStatus=1
  # look for errors at end of log
elif tac "$ciLog" | head | grep --color=always -Eie error; then
  printf \\n
  exitStatus=1
fi

printf \\n
msgln full logs
msgln ciLog:\\t\\t"$ciLog"

printf \\n
msgln took $(($(date +%s) - start))s

if [ "$exitStatus" != 0 ]; then
  printf "%b" "$LIB_COLOR_FAIL"
fi
