#!/usr/bin/env bash
#
# Initialize the Go tooling required.
#
set -euo pipefail

./contrib/go-get.sh v1.19.1 github.com/golangci/golangci-lint/cmd/golangci-lint
go get -u github.com/golang/dep/cmd/dep
