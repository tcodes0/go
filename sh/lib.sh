#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

########################################################
## this script is sourced by path from other scripts, ##
## careful if moving or renaming it                   ##
########################################################

set -euo pipefail
shopt -s globstar

export LIB_COLOR_PASS="\e[7;38;05;242m PASS \e[0m" LIB_COLOR_FAIL="\e[2;7;38;05;197;47m FAIL \e[0m"
export LIB_VISUAL_END="\e[0m" LIB_FORMAT_DIM="\e[2m" CHANGELOG_FILE="CHANGELOG.md"
# wrapper var to avoid macos sed incompatibilities
export SED=sed

# example: if macos;
macos() {
  [ "$(uname)" == "Darwin" ]
}

if macos; then
  SED=gsed
fi

# example: msgln hello world
msgln() {
  msg "$*\\n"
}

# internal, do not use
__log() {
  local level=$1 linenum=${2:?} msg=${*:3}

  if [ "$msg" ]; then
    echo -ne "$level ($0:$linenum) $msg\\n" >&2
  fi
}

# example: log some information
log() {
  __log INFO "$@"
}

# example: msg hello world
msg() {
  echo -ne "$*"
}

# output example: "23". Lines are terminal Y axis
currentTerminalLine() {
  # https://github.com/dylanaraps/pure-bash-bible#get-the-current-cursor-position
  IFS='[;' read -p $'\e[6n' -d R -rs _ currentLine _ _
  printf "%s" "$currentLine"
}

# example: requestedHelp "$*"
requestedHelp() {
  if ! [[ "$*" =~ -h|--help|help ]]; then
    return 1
  fi
}

# internal, do not use
__e() {
  local linenum=${1:?} funcname=$2 msg=ERROR

  if [ "${*:3}" ]; then
    msg=${*:3}
  fi

  echo -ne "$msg $0:$linenum ($funcname)" >&2
}

# usage: err $LINENO "message" (default message: error)
err() {
  __e "$1" "${FUNCNAME[1]}" "${*:2}"
}

#- - - - - - - - - - -

# usage: fatal $LINENO "message" (default message: error)
fatal() {
  __e "$1" "${FUNCNAME[1]}" "FATAL: ${*:2}"

  exit 1
}
