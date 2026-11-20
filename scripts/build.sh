#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o dist/touchbridge-linux-amd64 ./cmd/touchbridge
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o dist/touchbridge-linux-arm64 ./cmd/touchbridge
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o dist/touchbridge-windows-amd64.exe ./cmd/touchbridge
printf 'Built Linux amd64, Linux arm64 and Windows amd64 servers in %s/dist\n' "$ROOT"
