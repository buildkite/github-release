#!/bin/bash
set -eu

export GOOS=$1
export GOARCH=$2
export DISTFILE="dist/github-release-${GOOS}-${GOARCH}"

echo "Building github-release for $GOOS/$GOARCH 💨"
echo ""

rm -rf dist
mkdir -p dist

GOOS=darwin GOARCH=amd64 go build -o "${DISTFILE}" main.go
chmod +x "${DISTFILE}"
echo "👍 ${DISTFILE}"
echo ""

buildkite-agent artifact upload "${DISTFILE}"
