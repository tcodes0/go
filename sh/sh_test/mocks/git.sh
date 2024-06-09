#! /usr/bin/env bash
# Copyright 2024 Raphael Thomazella. All rights reserved.
# Use of this source code is governed by the BSD-3-Clause
# license that can be found in the LICENSE file and online
# at https://opensource.org/license/BSD-3-clause.

set -euo pipefail
shopt -s globstar

case "$1" in
tag)
  if [ "$2" == --list ]; then
    printf %b "$MOCK_TAG"
  fi
  ;;
show)
  printf %b "$MOCK_SHOW"
  ;;
*)
  printf %b "Command not mocked: $1"
  exit 1
  ;;
esac
