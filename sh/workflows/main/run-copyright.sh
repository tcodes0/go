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

usageExit() {
  msgln "Usage:"
  msgln "$0"
  msgln "REPORT=true $0"
  exit 1
}

### validation, input handling ###

### script ###

if requestedHelp "$*"; then
  usageExit
fi

report=""

if [ "${REPORT:-}" ]; then
  report="-report"
fi

go run ./cmd/copyright -globs "**/*.go, **/*/*.go, **/*/*/*.go, **/*.sh, **/*/*.sh, **/*/*.sh," -ignore "mock_.*, .local/.*" $report
