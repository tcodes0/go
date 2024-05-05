#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

start=$(date +%s)

import() {
  relativePath="go\/src\/sh\/shared-functions.sh"
  regex="(.*)\/go\/?.*" # \1 will capture base path
  functions=$(sed -E "s/$regex/\1\/$relativePath/g" <<<"$PWD")

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

currentLine=$(currentTerminalLine)
lastLine=$(tput lines)
firstColumn=0

if [ "$currentLine" -gt "$((lastLine - 10))" ]; then
  clear -x
  msg "running ci... (terminal cleared to make room for output)"
  currentLine=$(currentTerminalLine)
else
  msg "running ci..."
fi

linesPrinted=0
firstFailedJob=""
while ps -p $ciPid >/dev/null; do
  successToken="succeeded"
  failedToken="failed"

  # reset cursor
  tput cup "$currentLine" "$firstColumn"
  grepOut=$(grep -Eie "Job ($successToken|$failedToken)" "$ciLog" || true)
  linesPrinted=$(wc -l <<<"$grepOut" | sed 's/ .*//')

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

# move cursor to end of last loop output
tput cup "$((currentLine + linesPrinted))" "$firstColumn"

if [ -n "$firstFailedJob" ]; then
  printf "\n"
  grep --color=always -Eie "$firstFailedJob" "$ciLog"
  msg "above: logs for '$firstFailedJob'"
fi

printf "\n"
msg full logs
msg eventJson:\\t\\t"$eventJson"
msg ciLog:\\t\\t"$ciLog"

printf "\n"
msg took $(($(date +%s) - start))s
