#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck source=lib.sh
source "$PWD/sh/lib.sh"

##########################
### vars and functions ###
##########################

start=$(date +%s)
failed_jobs=()
passing_jobs=()

usage() {
  command cat <<-EOF
Usage:
Wrapper around act that runs ci workflows and presents output in a friendly way

$0
run with github actions pull-request event

$0 push
send a push event

$0 dispatch pizza
run release-pr workflow with tagprefix=pizza
EOF
}

# Description: Pushes a job into the global array
# Globals    : failed_jobs, passing_jobs
# Args       : 1=(pass | fail), 2=job name
# Sideeffects: Modifies global arrays
# Example    : push_job pass "deliver pizza"
push_job() {
  local type="$1" job="$2"

  job=$(echo -n "$job" | tr -s ' ')

  if [ "$type" == "fail" ]; then
    if [[ "${failed_jobs[*]}" != *${job}* ]]; then
      failed_jobs+=("$job")
    fi
  elif [[ "${passing_jobs[*]}" != *${job}* ]]; then
    passing_jobs+=("$job")
  fi
}

# Description: Reads log files to draw job progress
# Globals    : LIB_COLOR_PASS, LIB_COLOR_FAIL, LIB_FORMAT_DIM, LIB_VISUAL_END (lib.sh)
# Args       : 1=ci log file
# STDOUT     : Draws a job progress frame in terminal, manipulates cursor
# Example    : draw_progress /tmp/ci.log
draw_progress() {
  local success="succeeded" failed="failed" ci_log="$1"
  local reg_exp_job_status="\[([^]]+)\].*(succeeded|failed)"

  printf "\e[H" # move 1-1

  while read -r line; do
    if [[ ! "$line" =~ $reg_exp_job_status ]]; then
      continue
    fi

    local job="${BASH_REMATCH[1]}" status="${BASH_REMATCH[2]}"

    if [ "$status" == "$success" ]; then
      printf "%b %b%s%b\n" "$LIB_COLOR_PASS" "$LIB_FORMAT_DIM" "$job" "$LIB_VISUAL_END"
      push_job pass "$job"
    else
      printf "%b %b\n" "$LIB_COLOR_FAIL" "$job"
      push_job fail "$job"
    fi
  done < <(grep -Eie "Job ($success|$failed)" "$ci_log" || true)

  printf %ss\\n $(($(date +%s) - start))
}

# Description: Exits with failure if CI shouldn't run
# STDERR     : Might print errors
# Sideeffects: Might exits with failure
# Example    : validate
validate() {
  if [ "$(git status -s)" ]; then
    fatal $LINENO "please commit or stash all changes"
  fi

  if ! ping -c 1 1.1.1.1 >/dev/null; then
    fatal $LINENO "please check your internet connection"
  fi
}

# Description: Initializes ci log and event json file
# Args       : 1=github actions event, 2=tagprefix (see release-pr workflow)
# STDOUT     : event_type, log_file, event_json_file
# Returns    : event_type, log_file, event_json_file
# Sideeffects: Makes 2 temporary log files
# Example    : prepare_logs push pizza
prepare_logs() {
  local gitLocalBranch prJson eventJson pushJson releasePrJson
  local event="${1-}" tag_prefix="${2-}" event_json_file ciLog eventType

  gitLocalBranch=$(git branch --show-current)
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
  releasePrJson="
{
  \"inputs\": {
    \"tag_prefix\": \"$tag_prefix\",
    \"title\": \"CI Release\",
    \"url\": \"https://github.com/tcodes0/sh\"
  },
  \"local\": true
}
"

  case "$event" in
  "push")
    eventJson="$pushJson"
    eventType=push
    ;;
  "dispatch")
    eventJson="$releasePrJson"
    eventType=workflow_dispatch
    ;;
  *)
    eventJson="$prJson"
    eventType=pull_request
    ;;
  esac

  event_json_file=$(mktemp /tmp/ci-event-json-XXXXXX)
  ciLog=$(mktemp /tmp/ci-log-json-XXXXXX)

  printf "event json:" >"$ciLog"
  printf %s "$eventJson" >>"$ciLog"
  printf %s "$eventJson" >"$event_json_file"
  printf "%s %s %s" "$eventType" "$ciLog" "$event_json_file"
}

# Description: Print script usage
# Globals    : LIB_COLOR_FAIL (lib.sh), start, failed_jobs, passing_jobs
# Args       : 1=ci log file
# STDOUT     : Reports on the CI run
# Example    : report /tmp/ci.log
report() {
  local failed="" log="$1" minDurationSeconds=5
  msgln

  if [ $(($(date +%s) - start)) -le $minDurationSeconds ]; then
    tail "$log"
    failed=true
  elif [ "${#failed_jobs[@]}" != 0 ]; then
    msgln ${#failed_jobs[@]} jobs failed \(${#passing_jobs[@]} OK\)
    msgln see logs:

    for job in "${failed_jobs[@]}"; do
      msgln grep --color=always -Ee \""$job"\" "$log" \| less
    done

    failed=true
  elif [ ${#passing_jobs[@]} == 0 ]; then
    grep --color=always -Eie "error" "$log" || true
    msgln "error: no jobs succeeded"
    failed=true
    # look for errors at the end of log, fail if found
  elif tail "$log" | grep --color=always -Eie error; then
    failed=true
  fi

  msgln
  msgln full logs:\\t"$log"

  if [ "$failed" ]; then
    printf "%b" "$LIB_COLOR_FAIL"
  fi
}

# Description: Calls act with args, shows progress on screen, prints report
# Args       : Any
# STDOUT     : User feedback, progress and report on CI run
main() {
  if requested_help "$*"; then
    usage
    exit 1
  fi

  local type log_file event_json_file ci_command ci_command_args=() ci_pid

  validate
  read -rs type log_file event_json_file <<<"$(prepare_logs "$@")"

  ci_command="act"
  ci_command_args=("$type")
  ci_command_args+=(-e "$event_json_file")
  ci_command_args+=(-s GITHUB_TOKEN="$(gh auth token)")
  ci_command_args+=(--container-architecture linux/amd64)

  $ci_command "${ci_command_args[@]}" >>"$log_file" 2>&1 || true &
  ci_pid=$!

  printf "\e[H\e[2J" # move 1-1, clear whole screen
  msgln "    running ci..."
  trap 'msgln \\nsee partial ci log at $log_file' INT

  while ps -p "$ci_pid" >/dev/null; do
    draw_progress "$log_file"
    sleep 1
  done

  # catch status of last job that could have been missed by loop
  draw_progress "$log_file"
  report "$log_file"
}

##############
### script ###
##############

main "$@"
