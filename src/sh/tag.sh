#! /usr/bin/env bash

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/src/sh/lib.sh"

if [ -z "${EXEC_GIT_READ:-}" ]; then
  EXEC_GIT_READ=git
fi
if [ -z "${EXEC_GIT_WRITE:-}" ]; then
  EXEC_GIT_WRITE="git"
fi

### vars and functions ###

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
  ["commit"]="c"
)

declare -rA optsHelp=(
  [${opts["pre"]}]="bool; start a new pre-release from 1"
  [${opts["dry"]}]="bool; dry-run, print commands that would be executed"
  [${opts["commit"]}]="string; commit hash to tag"
)

declare -A optValue=(
  # defaults
  ["pre"]=""
  ["dry"]=""
  ["commit"]="HEAD"
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

# composes a tag
# example: tag 1 2 3 pre 4 outputs v1.2.3-pre4
tag() {
  if [ -n "${4:-}" ]; then
    printf 'v%s.%s.%s-pre%s' "$1" "$2" "$3" "$5"
  else
    printf 'v%s.%s.%s' "$1" "$2" "$3"
  fi
}

# increments the major version
# example: major 1 2 3 pre 4 outputs v2.0.0
major() {
  if [ "${optValue["pre"]}" ]; then
    printf %s "$(tag "$((1 + 1))" 0 0 pre 1)"
  else
    printf %s "$(tag "$((1 + 1))" 0 0)"
  fi
}

# increments the minor version
# example: minor 1 2 3 outputs v1.3.0
minor() {
  if [ "${optValue["pre"]}" ]; then
    printf %s "$(tag "$1" "$(($2 + 1))" 0 pre 1)"
  else
    printf %s "$(tag "$1" "$(($2 + 1))" 0)"
  fi
}

# increments the patch version
# example: bump 1 2 3 outputs v1.2.4
bump() {
  if [ -n "$4" ]; then
    printf %s "$(tag "$1" "$2" "$3" pre "$(($5 + 1))")"
    return
  fi

  if [ "${optValue["pre"]}" ]; then
    printf %s "$(tag "$1" "$2" "$(($3 + 1))" pre 1)"
  else
    printf %s "$(tag "$1" "$2" "$(($3 + 1))")"
  fi
}

# adds a tag to the commit specified
# example: addTag v1.2.3
addTag() {
  $EXEC_GIT_WRITE tag "$1" "${optValue["commit"]}" || msgExit "git tag failed"

  local formatShortHashMessageTags="%h %s (%D)"
  $EXEC_GIT_READ show --format="$formatShortHashMessageTags" "${optValue["commit"]}" | head -1 | grep --color=auto -Ee 'tag:[^,]+'
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
while getopts "${opts["pre"]}${opts["dry"]}${opts["commit"]}:" opt; do
  #   echo "opt: $opt", "OPTARG: ${OPTARG:-}"
  case $opt in
  "${opts["pre"]}")
    optValue["pre"]=true
    ;;

  "${opts["dry"]}")
    optValue["dry"]=true
    ;;

  "${opts["commit"]}")
    optValue["commit"]="$OPTARG"
    ;;

  \?)
    usageExit "Invalid option: $OPTARG"
    ;;
  esac
done

if [ "${optValue["dry"]}" ]; then
  EXEC_GIT_WRITE="echo dry run: git"
fi

### script ###

IFS=$'\n' read -rd "$CHAR_CARRIG_RET" -a tags < <(
  set +e # flaky for some reason
  $EXEC_GIT_READ tag --list --sort=-refname | head
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

case $commandArg in
"${commands["major"]}")
  addTag "$(major "$tagMajor" "$tagMinor" "$tagPatch" "$tagPre" "$tagPreVersion")"
  ;;
"${commands["minor"]}")
  addTag "$(minor "$tagMajor" "$tagMinor" "$tagPatch" "$tagPre" "$tagPreVersion")"
  ;;
"${commands["bump"]}")
  addTag "$(bump "$tagMajor" "$tagMinor" "$tagPatch" "$tagPre" "$tagPreVersion")"
  ;;
esac
