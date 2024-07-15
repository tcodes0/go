#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck disable=SC1091
source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

### vars and functions ###

### script ###

if ! [[ $(git log --oneline --decorate | head -1) =~ chore:\ release ]]; then
  log "main head not a release commit"
  exit 0
fi

if ! [[ $(head -1 "$CHANGELOG_FILE") =~ \#\ ([[:alnum:]]+):\ (v.+\..+\.[[:digit:]]+) ]]; then
  err "malformed changelog head"
  exit 1
fi

tag="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
msgln "$tag"
git tag "$tag" HEAD
git push origin --tags
