#! /usr/bin/env bash

set -euo pipefail
shopt -s globstar

echo "GO_MOD=${GO_MOD}"
goModChanged="true"
if [ -z "$GO_MOD" ]; then
  goModChanged=""
fi

echo "CONFIG=${CONFIG}"
configChanged="true"
if [ -z "$CONFIG" ]; then
  configChanged=""
fi

echo "DOC=${DOC}"
docChanged="true"
if [ -z "$DOC" ]; then
  docChanged=""
fi

echo "SHELL=${SHELL}"
shellChanged="true"
if [ -z "$SHELL" ]; then
  shellChanged=""
fi

echo "GO_PKG_HTTPFLUSH=${GO_PKG_HTTPFLUSH}"
goPkgHttpflushChanged="true"
if [ -z "$GO_PKG_HTTPFLUSH" ]; then
  goPkgHttpflushChanged=""
fi

IFS=" " read -r -a anyGoPkgChanged <<<"$goPkgHttpflushChanged"

echo "goModChanged=$goModChanged"                   >> "$GITHUB_OUTPUT"
echo "goModChanged=$goModChanged"
echo "configChanged=$configChanged"                 >> "$GITHUB_OUTPUT"
echo "configChanged=$configChanged"
echo "docChanged=$docChanged"                       >> "$GITHUB_OUTPUT"
echo "docChanged=$docChanged"
echo "shellChanged=$shellChanged"                   >> "$GITHUB_OUTPUT"
echo "shellChanged=$shellChanged"
echo "goPkgHttpflushChanged=$goPkgHttpflushChanged" >> "$GITHUB_OUTPUT"
echo "goPkgHttpflushChanged=$goPkgHttpflushChanged"
echo "anyGoPkgChanged=${anyGoPkgChanged[*]}"        >> "$GITHUB_OUTPUT"
echo "anyGoPkgChanged=${anyGoPkgChanged[*]}"