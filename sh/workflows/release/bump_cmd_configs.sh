#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar nullglob

while read -r line; do
  if [[ "$line" =~ cmd\/v([[:digit:]]+\.[[:digit:]]+.[[:digit:]]) ]]; then
    find cmd -name config.yml -exec yq eval ".version = ${BASH_REMATCH[1]}" -i {} \;
    break
  fi
done <"$TAGS_FILE"
