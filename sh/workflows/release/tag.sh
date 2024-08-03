#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar
# shellcheck source=../../lib.sh
source "$PWD/sh/lib.sh"
trap 'err $LINENO' ERR

### vars and functions ###

release_re="chore:[ ]release"
changelog_head_re="#[ ]([[:alnum:]]+):[ ](v.+\..+\.[[:digit:]]+)"

### script ###

if ! [[ $(git log --oneline --decorate | head -1) =~ $release_re ]]; then
  log "main head not a release commit"
  exit 0
fi

if ! [[ $(head -1 "$CHANGELOG_FILE") =~ $changelog_head_re ]]; then
  err "malformed changelog head, expected '# something: v0.1.2'"
  exit 1
fi

tag="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
msgln "$tag"
git tag "$tag" HEAD
git push origin --tags
