#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

export PASS_GREEN="\e[7;38;05;029m PASS \e[0m"
export FAIL_RED="\e[2;7;38;05;197;47m FAIL \e[0m"
export COLOR_DIM="\e[2m"
export COLOR_END="\e[0m"

msg() {
  echo -e "> $*"
}

msgExit() {
  msg "$*"
  return 1
}

requireGitClean() {
  if [ -n "$(git diff --exit-code)" ]; then
    msgExit "There are uncommitted changes, please commit or stash"
  fi
}

currentTerminalLine() {
  # https://github.com/dylanaraps/pure-bash-bible#get-the-current-cursor-position
  IFS='[;' read -p $'\e[6n' -d R -rs _ currentLine _ _
  printf "$currentLine"
}

# mac os notes
# mktemp needs placeholder X's at the end of the string
# avoid omitting -e to sed to pass expression
# sleep takes only a number, seconds, as argument, contrary to linux sleep that can take "1s"
