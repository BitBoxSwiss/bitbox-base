#!/usr/bin/env bash
#
# Initialize the Go tooling required.
#
set -euo pipefail

curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.19.1
GO111MODULE=off go get -u github.com/vektra/mockery/cmd/mockery
