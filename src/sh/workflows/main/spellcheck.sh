#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

npm install -g "cspell@$VERSION"
cspell "$FILES"
