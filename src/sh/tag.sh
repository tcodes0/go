#! /usr/bin/env bash

### options and imports ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

### vars and functions ###

if [ -z "${EXEC_GIT_READ:-}" ]; then
  EXEC_GIT_READ=git
fi
if [ -z "${EXEC_GIT_WRITE:-}" ]; then
  EXEC_GIT_WRITE="git"
fi

declare -rA commands=(
  ["major"]="major"
  ["minor"]="minor"
  ["bump"]="bump"
)

declare -rA commandsHelp=(
  [${commands["major"]}]="increment the major version"
  [${commands["minor"]}]="increment the minor version"
  [${commands["bump"]}]="increment the patch version or pre-release version"
)

declare -rA opts=(
  ["pre"]="p"
  ["dry"]="n"
)

declare -rA optsHelp=(
  [${opts["pre"]}]="bool; start a new pre-release from 1"
  [${opts["dry"]}]="bool; dry-run, print commands that would be executed"
)

declare -A optValue=(
  # defaults
  ["pre"]=""
  ["dry"]=""
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
shift

if ! [[ " ${commands[*]} " =~ $commandArg ]]; then
  usageExit "Invalid command: $commandArg"
fi

OPTIND=1
while getopts "${opts["pre"]}${opts["dry"]}" opt; do
  #   echo "opt: $opt", "OPTARG: ${OPTARG:-}"
  case $opt in
  "${opts["pre"]}")
    optValue["pre"]=true
    ;;
  "${opts["dry"]}")
    optValue["dry"]=true
    ;;
  \?)
    usageExit "Invalid option: $OPTARG"
    ;;
  esac
done

if [ "${optValue["dry"]}" ]; then
  EXEC_GIT_WRITE="echo git"
fi

### script ###

IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a tags < <(
  set +e # flaky for some reason
  $EXEC_GIT_READ tag --list --sort=-refname | head
  printf %b "$CHAR_CARRIG_RET"
)
IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a logs < <(
  set +e # flaky for some reason
  $EXEC_GIT_READ log --oneline --decorate | head
  printf %b "$CHAR_CARRIG_RET"
)

latestTag="${tags[0]}"
regExpSemVerPre="v?([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+)-?(pre)?([[:digit:]]*)?"

if ! [[ "$latestTag" =~ $regExpSemVerPre ]]; then
  msgExit "parse fail, tag: $latestTag, regExp: $regExpSemVerPre"
fi

tagMajor="${BASH_REMATCH[1]}"
tagMinor="${BASH_REMATCH[2]}"
tagPatch="${BASH_REMATCH[3]}"
tagPre="${BASH_REMATCH[4]:-}"
tagPreVersion="${BASH_REMATCH[5]:0}"

# echo "latestTag: $latestTag"
# echo "tagMajor: $tagMajor"
# echo "tagMinor: $tagMinor"
# echo "tagPatch: $tagPatch"
# echo "tagPre: $tagPre"
# echo "tagPreVersion: $tagPreVersion"

next=""

major() {
  if [ "${optValue["pre"]}" ]; then
    next=$(tag "$((tagMajor + 1))" 0 0 pre 1)
  else
    next=$(tag "$((tagMajor + 1))" 0 0)
  fi
}

minor() {
  if [ "${optValue["pre"]}" ]; then
    next=$(tag "$tagMajor" "$((tagMinor + 1))" 0 pre 1)
  else
    next=$(tag "$tagMajor" "$((tagMinor + 1))" 0)
  fi
}

bump() {
  if [ -n "$tagPre" ]; then
    next=$(tag "$tagMajor" "$tagMinor" "$tagPatch" pre "$((tagPreVersion + 1))")
    return
  fi

  if [ "${optValue["pre"]}" ]; then
    next=$(tag "$tagMajor" "$tagMinor" "$((tagPatch + 1))" pre 1)
  else
    next=$(tag "$tagMajor" "$tagMinor" "$((tagPatch + 1))")
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

$EXEC_GIT_WRITE tag "$next" || msgExit "git tag failed"
$EXEC_GIT_READ show --decorate HEAD | head -1
printf "tagged with %s\n" "$next"
