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

flagVerbose="-v"
ciCommand="act"

usageExit() {
  msg "Usage: $0 [$flagVerbose]"
  msg "$flagVerbose: print to console"
  exit 1
}

if [ $# == 1 ] && [ "$1" != "$flagVerbose" ]; then
  msg "Invalid argument: $1"
  usageExit
fi

if [ $# -gt 1 ]; then
  msg "Invalid number of arguments: $# ($*)"
  usageExit
fi

requireGitClean

eventJson=$(mktemp)
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

log=$(mktemp /tmp/ci-pull-request-XXXXXX.log)

verbose=""
if [[ "$*" == *${flagVerbose}* ]]; then
  verbose="true"
fi

ciCommandArgs=(-e "$eventJson")
ciCommandArgs+=(-s GITHUB_TOKEN="$(gh auth token)")

# allow ci and grep to fail without killing script
set +e
if [ -n "$verbose" ]; then
  $ciCommand "${ciCommandArgs[@]}" 2>&1 | tee "$log"
else
  msg "running ci..."
  $ciCommand "${ciCommandArgs[@]}" 2>&1 | tee "$log" >/dev/null
  msg "ci exited with $?"
fi


if [ -z "$verbose" ]; then
  successToken="Job succeeded"
  grep --color=always -Eie "$successToken" "$log" || true

  printf "\n"
  failedToken="Job failed"
  grep --color=always -Eie "$failedToken" "$log" || true
  regex="1s/\[([^]]+)\].*/\1/p"
  firstFailedJobName=$(grep "$failedToken" "$log" | sed -nE "$regex")

  if [ -n "$firstFailedJobName" ]; then
    printf "\n"
    msg "logs for '$firstFailedJobName':"
    grep --color=always -Eie "$firstFailedJobName" "$log" || true
  fi
fi

printf "\n"
msg run variables
msg eventJson:\\t\\t"$eventJson"
msg gitLocalBranch:\\t"$gitLocalBranch"
msg log:\\t\\t\\t"$log"

printf "\n"
msg took $(($(date +%s) - start))s
