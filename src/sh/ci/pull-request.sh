#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

start=$(date +%s)

usageExit() {
  msg "Usage: $0 "
  exit 1
}

if [ $# != 0 ]; then
  msg "Invalid argument: $1"
  usageExit
fi

requireGitClean

eventJson=$(mktemp /tmp/ci-event-json-XXXXXX)
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
  },
  \"local\": true
}
" >"$eventJson"

ciCommand="act"
ciCommandArgs=(-e "$eventJson")
ciCommandArgs+=(-s GITHUB_TOKEN="$(gh auth token)")
ciCommandArgs+=(--container-architecture linux/amd64)
ciLog=$(mktemp /tmp/ci-log-json-XXXXXX)

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

firstFailedJob=""
hasSuccessfulJob=""
iterations=0
while ps -p $ciPid >/dev/null; do
  successToken="succeeded"
  failedToken="failed"

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

  iterations=$((iterations + 1))
  sleep 1
done

exitStatus=0
somethingWrong=5
if [ "$iterations" -lt "$somethingWrong" ]; then
  printf "\n"
  head "$ciLog"
  exitStatus=1
fi

if [ "$exitStatus" == 0 ] && [ -n "$firstFailedJob" ]; then
  printf "\n"
  grep --color=always -Eie "$firstFailedJob" "$ciLog" || true
  msg "above: logs for '$firstFailedJob'"
  exitStatus=1
fi

if [ "$exitStatus" == 0 ] && [ -z "$hasSuccessfulJob" ]; then
  printf "\n"
  grep --color=always -Eie "error" "$ciLog" || true
  msg "error: no jobs succeeded"
  exitStatus=1
fi

if [ "$exitStatus" == 0 ]; then
  # look for errors at end of log
  tac "$ciLog" | head | grep --color=always -Eie error || true
  exitStatus=1
fi

printf "\n"
msg full logs
msg eventJson:\\t\\t"$eventJson"
msg ciLog:\\t\\t"$ciLog"

printf "\n"
msg took $(($(date +%s) - start))s

if [ "$exitStatus" != 0 ]; then
  printf "%b" "$FAIL_RED"
fi

exit "$exitStatus"
