#! /usr/bin/env bash

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
