#!/bin/bash
set -eu

echo "Building github-release 💨"
echo ""

rm -rf dist
mkdir -p dist

echo "Compiling for OSX"
docker-compose env GOOS=darwin GOARCH=amd64 go build -o dist/github-release-darwin-amd64 main.go
chmod +x dist/github-release-darwin-amd64
echo "👍  dist/github-release-darwin-amd64"
echo ""

echo "Compiling for Linux"
GOOS=linux GOARCH=amd64 go build -o dist/github-release-linux-amd64 main.go
chmod +x dist/github-release-linux-amd64
echo "👍  dist/github-release-linux-amd64"

echo ""
echo "All done! ✅"
