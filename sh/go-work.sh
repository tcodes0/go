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
pkgs=$(packages)
formattedPkgs=""

for pkg in $pkgs; do
  formattedPkgs=${formattedPkgs}$(printf %s "	$pkg\\n")
done

printf %b "$goVersion

use (
	.
$formattedPkgs
)
" >go.work
