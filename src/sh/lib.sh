#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

export PASS_GREEN="\e[7;38;05;029m PASS \e[0m"
export FAIL_RED="\e[2;7;38;05;197;47m FAIL \e[0m"
export COLOR_DIM="\e[2m"
export COLOR_END="\e[0m"

# msg hello world
msg() {
  echo -e "> $*"
}

# msgExit could not find the file
msgExit() {
  msg "$*"
  return 1
}

# requireGitClean please commit changes to avoid losing work
requireGitClean() {
  message="${*:-There are uncommitted changes, please commit or stash}"

  if [ -n "$(git diff --exit-code)" ]; then
    msgExit "$message"
  fi
}

# output example: "23". Lines are terminal Y axis
currentTerminalLine() {
  # https://github.com/dylanaraps/pure-bash-bible#get-the-current-cursor-position
  IFS='[;' read -p $'\e[6n' -d R -rs _ currentLine _ _
  printf "%s" "$currentLine"
}

# requireInternet Internet required to fetch dependencies
requireInternet() {
  pingPal="1.1.1.1" # cloudflare
  message="${*:-Internet required}"

  if ! ping -c 1 "$pingPal" &>/dev/null; then
    msgExit "$message"
  fi
}
