#! /usr/bin/env bash

set -euo pipefail

linters=(
  defers
  fieldalignment
  findcall
  httpmux
  ifaceassert
  lostcancel
  nilness
  shadow
  stringintconv
  unmarshal
  unusedresult
)

for linter in "${linters[@]}"; do
  $linter --fix "$PWD/$1"
done
