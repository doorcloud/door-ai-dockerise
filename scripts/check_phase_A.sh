#!/usr/bin/env bash
set -euo pipefail
set -x

echo "▶ go vet";          go vet ./...
echo "▶ unit tests";      go test ./...
echo "▶ go mod tidy";     go mod tidy
git diff --exit-code go.{mod,sum} || { echo "tidy changed module files"; exit 1; }
echo "▶ build cli";       go build -o /tmp/dockergen ./cmd/dockergen
echo "✓ Phase-A OK" 