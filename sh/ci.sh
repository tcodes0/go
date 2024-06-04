#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

start=$(date +%s)

usageExit() {
  msgln "Usage: $0 "
  exit 1
}

firstFailedJob=""
hasSuccessfulJob=""
printJobProgress() {
  successToken="succeeded"
  failedToken="failed"
  ciLog="$1"

  tput rc
  grepOut=$(grep -Eie "Job ($successToken|$failedToken)" "$ciLog" || true)
  regExpAfterSpace=" .*"
  linesPrinted=$(wc -l <<<"$grepOut" | sed -e "s/$regExpAfterSpace//")

  if [ "$linesPrinted" != 0 ]; then
    while read -r line; do
      if [ -z "$line" ]; then
        continue
      fi

      if [[ "$line" =~ \[([^]]+)\].*(succeeded|failed) ]]; then
        job="${BASH_REMATCH[1]}"
        status="${BASH_REMATCH[2]}"

        if [ "$status" == "$successToken" ]; then
          printf "%b %b%s%b\n" "$COLOR_PASS" "$FORMAT_DIM" "$job" "$VISUAL_END"
          hasSuccessfulJob=true
        else
          printf "%b %b\n" "$COLOR_FAIL" "$job"
          if [ -z "$firstFailedJob" ]; then
            firstFailedJob="$job"
          fi
        fi
      fi
    done <<<"$grepOut"
  fi
}

if [ $# != 0 ]; then
  msgln "Invalid argument: $1"
  usageExit
fi

eventJson=$(mktemp /tmp/ci-event-json-XXXXXX)
gitLocalBranch=$(git branch --show-current)

printf "
{
  \"pull_request\": {
    \"title\": \"feat(ci): add PR title to act event\",
    \"head\": {
      \"ref\": \"%s\"
    },
    \"base\": {
      \"ref\": \"main\"
    }
  },
  \"local\": true
}
" "$gitLocalBranch" >"$eventJson"

ciCommand="act"
ciCommandArgs=(-e "$eventJson")
ciCommandArgs+=(-s GITHUB_TOKEN="$(gh auth token)")
ciCommandArgs+=(--container-architecture linux/amd64)
ciLog=$(mktemp /tmp/ci-log-json-XXXXXX)

$ciCommand "${ciCommandArgs[@]}" >"$ciLog" 2>&1 || true &
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
msgln eventJson:\\t\\t"$eventJson"
msgln ciLog:\\t\\t"$ciLog"

printf \\n
msgln took $(($(date +%s) - start))s

if [ "$exitStatus" != 0 ]; then
  printf "%b" "$COLOR_FAIL"
fi
