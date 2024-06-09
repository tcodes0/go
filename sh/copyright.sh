#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

### options, imports, mocks ###

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"

### vars and functions ###

boiler=$(cat ./sh/copyright-header.txt)
regExpLineStart="^"
regExpGoCommentSpace="\/\/ "
regExpShCommentSpace="# "
boilerGo=$(_sed --regexp-extended -e "s/$regExpLineStart/$regExpGoCommentSpace/" <<<"$boiler")
boilerSh=$(_sed --regexp-extended -e "s/$regExpLineStart/$regExpShCommentSpace/" <<<"$boiler")
globs=('*.go' '*.sh')
ignoreRegExps="/?mock_.*|.local/.*"
missing=()

usageExit() {
  msgln "Check and fix missing boilerplate header in files."
  msgln "Usage: $0       fails if files are missing copyright header and prints files"
  msgln "Usage: $0 -fix  applies header to files"
  exit 1
}

filesMissingBoilerplate() {
  glob="$1"
  paths=$(find . -name "$glob" -type f)

  while read -r path; do
    if [[ "$path" =~ $ignoreRegExps ]]; then
      debugMsg ignore "$path"
      continue
    fi

    if head "$path" | grep -q 'Copyright'; then
      debugMsg "ok $path"
    else
      missing+=("$path")
    fi
  done <<<"$paths"
}

fixBoilerplate() {
  path="$1"
  newFile=""

  if [[ "$path" == *.go ]]; then
    newFile="${boilerGo}\n\n$(cat "$path")\n"
  elif [[ "$path" == *.sh ]]; then
    read -rs lastLine _ < <(wc -l "$path")
    shebang=$(_sed -n '1p' "$path")
    newFile="${shebang}\n${boilerSh}\n$(_sed -n "2,${lastLine}p" "$path")\n"
  fi

  printf "%b" "$newFile" >"$path"
}

### validation, input handling ###

if requestedHelp "$*"; then
  usageExit
fi

### script ###

for glob in "${globs[@]}"; do
  filesMissingBoilerplate "$glob"
done

for path in "${missing[@]}"; do
  msgln "$path"
done

if [ ${#missing[@]} == 0 ]; then
  exit
fi

if [ "${1:-}" == -fix ]; then
  for path in "${missing[@]}"; do
    fixBoilerplate "$path"
  done
else
  exit 1
fi
