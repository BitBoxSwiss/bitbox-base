#!/usr/bin/env bash
#
# Initialize the Go tooling required.
#
set -euo pipefail

./scripts/go-get.sh v1.16.0 github.com/golangci/golangci-lint/cmd/golangci-lint
go get -u github.com/golang/dep/cmd/dep
