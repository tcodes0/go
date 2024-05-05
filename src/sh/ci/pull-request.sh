#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

start=$(date +%s)

import() {
  relativePath="go\/src\/sh\/lib.sh"
  regExpBasePath="(.*)\/go\/?.*"
  functions=$(sed -E "s/$regExpBasePath/\1\/$relativePath/g" <<<"$PWD")

  # shellcheck disable=SC1090
  source "$functions"
}

import

usageExit() {
  msg "Usage: $0 "
  exit 1
}

if [ $# != 0 ]; then
  msg "Invalid argument: $1"
  usageExit
fi

requireGitClean

eventJson=$(mktemp /tmp/ci-event-XXXXXX.json)
gitLocalBranch=$(git branch --show-current)

printf %s "
{
  \"pull_request\": {
    \"head\": {
      \"ref\": \"$gitLocalBranch\"
    },
    \"base\": {
      \"ref\": \"main\"
    }
  }
}
" >"$eventJson"

ciCommand="act"
ciCommandArgs=(-e "$eventJson")
ciCommandArgs+=(-s GITHUB_TOKEN="$(gh auth token)")
ciLog=$(mktemp /tmp/ci-XXXXXX.log)

$ciCommand "${ciCommandArgs[@]}" 2>&1 | tee "$ciLog" >/dev/null || true &
ciPid=$!

lastLine=$(tput lines)

if [ "$(currentTerminalLine)" -gt "$((lastLine - 10))" ]; then
  clear -x
  msg "running ci... (terminal cleared to make room for output)"
else
  msg "running ci..."
fi

tput sc

linesPrinted=0
firstFailedJob=""
hasSuccessfulJob=""
while ps -p $ciPid >/dev/null; do
  successToken="succeeded"
  failedToken="failed"

  tput rc
  grepOut=$(grep -Eie "Job ($successToken|$failedToken)" "$ciLog" || true)
  regExpAfterSpace=" .*"
  linesPrinted=$(wc -l <<<"$grepOut" | sed "s/$regExpAfterSpace//")

  if [ "$linesPrinted" != 0 ]; then
    while read -r line; do
      if [ -z "$line" ]; then
        continue
      fi

      if [[ "$line" =~ \[([^]]+)\].*(succeeded|failed) ]]; then
        job="${BASH_REMATCH[1]}"
        status="${BASH_REMATCH[2]}"

        if [ "$status" == "$successToken" ]; then
          printf "%b %b%s%b\n" "$PASS_GREEN" "$COLOR_DIM" "$job" "$COLOR_END"
          hasSuccessfulJob=true
        else
          printf "%b %b\n" "$FAIL_RED" "$job"
          if [ -z "$firstFailedJob" ]; then
            firstFailedJob="$job"
          fi
        fi
      fi
    done <<<"$grepOut"
  fi

  sleep 1s
done

if [ -n "$firstFailedJob" ]; then
  printf "\n"
  grep --color=always -Eie "$firstFailedJob" "$ciLog"
  msg "above: logs for '$firstFailedJob'"
fi

if [ -z "$hasSuccessfulJob" ]; then
  printf "\n"
  grep --color=always -Eie "error" "$ciLog"
  msg "error: no jobs suceeded"
fi

printf "\n"
msg full logs
msg eventJson:\\t\\t"$eventJson"
msg ciLog:\\t\\t"$ciLog"

printf "\n"
msg took $(($(date +%s) - start))s

if [ -z "$hasSuccessfulJob" ]; then
  exit 1
fi
