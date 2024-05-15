#! /usr/bin/env bash

### options and imports ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

declare -A commands=(
  ["major"]="major"
  ["minor"]="minor"
  ["bump"]="bump"
)

declare -A commandsHelp=(
  [${commands["major"]}]="increment the major version"
  [${commands["minor"]}]="increment the minor version"
  [${commands["bump"]}]="increment the patch version or pre-release version"
)

declare -A opts=(
  ["pre"]="p"
)

declare -A optsHelp=(
  [${opts["pre"]}]="bool; create a pre-release version"
)

usageExit() {
  commandsHelpInfo=$(
    IFS=\|
    printf "%s" "${!commandsHelp[*]}"
  )
  optsHelpInfo=$(
    IFS=-
    printf "%s" -"${!optsHelp[*]}"
  )

  msg "$*\n"
  msg "Usage: $0 [$commandsHelpInfo] [$optsHelpInfo]"

  printf "\n"
  for command in "${!commandsHelp[@]}"; do
    msg "$command:  ${commandsHelp[$command]}"
  done

  printf "\n"
  for opt in "${!optsHelp[@]}"; do
    msg "-$opt:  ${optsHelp[$opt]}"
  done

  exit 1
}

# example: tag 1 2 3 pre 4 outputs v1.2.3-pre4
tag() {
  if [ -n "${4:-}" ]; then
    printf 'v%s.%s.%s-pre%s' "$1" "$2" "$3" "$5"
  else
    printf 'v%s.%s.%s' "$1" "$2" "$3"
  fi
}

### validation, input handling ###

if [ $# -lt 1 ]; then
  usageExit "Invalid number of arguments $# ($*)"
fi

commandArg=$1

### script ###

IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a tags < <(
  git tag --list --sort=-refname | head
  printf %b "$CHAR_CARRIG_RET"
)
IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a logs < <(
  git log --oneline --decorate | head
  printf %b "$CHAR_CARRIG_RET"
)

latestTag="${tags[0]}"
regExpSemVerPre="v?([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+)-?(pre)?([[:digit:]]*)?"

[[ "$latestTag" =~ $regExpSemVerPre ]]
major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"
patch="${BASH_REMATCH[3]}"
preRelease="${BASH_REMATCH[4]:-}"
PreReleaseVer="${BASH_REMATCH[5]:-}"

# echo "latestTag: $latestTag"
# echo "major: $major"
# echo "minor: $minor"
# echo "patch: $patch"
# echo "preRelease: $preRelease"
# echo "PreReleaseVer: $PreReleaseVer"

next=""

major() {
  next=$(tag "$((major + 1))" 0 0)
}

minor() {
  next=$(tag "$major" "$((minor + 1))" 0)
}

bump() {
  if [ -z "$preRelease" ]; then
    next=$(tag "$major" "$minor" "$((patch + 1))")
  else
    next=$(tag "$major" "$minor" "$patch" pre "$((PreReleaseVer + 1))")
  fi
}

case $commandArg in
"${commands["major"]}")
  major
  ;;
"${commands["minor"]}")
  minor
  ;;
"${commands["bump"]}")
  bump
  ;;
esac

printf "current\t%s\n" "$latestTag"
printf "next\t%s\n" "$next"
