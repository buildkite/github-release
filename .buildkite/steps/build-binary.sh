#!/bin/bash
set -eu

export GOOS=$1
export GOARCH=$2
export DISTFILE="dist/github-release-${GOOS}-${GOARCH}"

go_version="1.9.2"
go_pkg="github.com/buildkite/github-release"

rm -rf dist
mkdir -p dist

go_build_in_docker() {
  docker run \
    -v "${PWD}:/go/src/${go_pkg}" \
    -w "/go/src/${go_pkg}" \
    -e "GOOS=${GOOS}" -e "GOARCH=${GOARCH}" -e "CGO_ENABLED=0" \
    --rm "golang:${go_version}" \
    go build "$@"
}

echo "+++ Building github-release for $GOOS/$GOARCH with golang:${go_version} :golang:"

go_build_in_docker -a -tags netgo -ldflags '-w' -o "${DISTFILE}" main.go
file "${DISTFILE}"

chmod +x "${DISTFILE}"
echo "👍 ${DISTFILE}"

buildkite-agent artifact upload "${DISTFILE}"
