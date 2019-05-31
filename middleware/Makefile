.DEFAULT_GOAL := aarch64
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/..

check-env:
	./scripts/check-env.sh "$(REPO_ROOT)"

ci:
	./scripts/ci.sh
envinit:
	# TODO(hkjn): Make this rule more robust, or remove it if there's no clear
	# need for it.
	./scripts/go-get.sh v1.16.0 github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u github.com/golang/dep/cmd/dep
native: check-env
	go build -o $(REPO_ROOT)/build/base-middleware ./src/
aarch64: check-env
	GOARCH=arm64 go build -o $(REPO_ROOT)/build/base-middleware ./src/
