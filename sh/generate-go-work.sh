#! /usr/bin/env bash

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

for mod in $mods; do
  formattedMods=${formattedMods}$(printf %s "	$mod\\n")
done

printf %b "// generated do not edit.
$goVersion

use (
	.
$formattedMods
)
" >go.work
