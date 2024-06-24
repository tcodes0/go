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
export LIB_VISUAL_END="\e[0m" LIB_FORMAT_DIM="\e[2m"

# example: msgln hello world
msgln() {
  msg "$*\\n"
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

# example: if macos;
macos() {
  ([ "$(uname)" == "Darwin" ] && true)
}

# wrapper to avoid macos sed incompatibilities
_sed() {
  if macos; then
    gsed "$@"
    return
  fi

  sed "$@"
}

# err $LINENO "message" (default message: error)
err() {
  linenum=$1
  msg=error

  if [ "${*:2}" ]; then
    msg=${*:2}
  fi

  echo "$msg: $0":"$linenum" \("${FUNCNAME[1]}"\) >&2
}