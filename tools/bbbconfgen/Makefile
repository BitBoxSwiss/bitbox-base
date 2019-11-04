.DEFAULT_GOAL := aarch64
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..

check-go-env:
	@echo "Checking that environment supports Go builds.."
	@$(REPO_ROOT)/contrib/check-go-env.sh "$(REPO_ROOT)"

native: check-go-env
	go install $(REPO_ROOT)/tools/bbbconfgen

aarch64: check-go-env
	GOARCH=arm64 go build $(REPO_ROOT)/tools/bbbconfgen
	cp $(REPO_ROOT)/tools/bbbconfgen/bbbconfgen $(REPO_ROOT)/bin/go/
