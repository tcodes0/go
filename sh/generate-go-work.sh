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

parseGoVersion() {
  while read -r line; do
    if [[ $line =~ ^go ]]; then
      printf %s "$line"
      break
    fi
  done <go.mod
}

### script ###

goVersion=$(parseGoVersion)
mods=$(findModules)
formattedMods=""
regexpPathHasSlash="(.*[[:alnum:]])\/([[:alnum:]].*)"

for mod in $mods; do
  if [ "${IGNORE:-}" ]; then
    if [[ $mod =~ $IGNORE ]]; then
      continue
    fi
  fi

  if [[ $mod =~ $regexpPathHasSlash ]]; then
    # use base module name
    # if deeply nested this check can be used recursively until no more slashes are found
    mod=${BASH_REMATCH[1]}
  fi

  formattedMods=${formattedMods}$(printf %s "	$mod\\n")
done

printf %b "// generated do not edit.
$goVersion

use (
	.
$formattedMods
)
" >go.work
