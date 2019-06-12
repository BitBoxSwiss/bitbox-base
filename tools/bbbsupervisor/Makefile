.DEFAULT_GOAL := aarch64
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..

build-target-exists:
	@mkdir -p $(REPO_ROOT)/build

check-go-env: build-target-exists
	@echo "Checking that environment supports Go builds.."
	@$(REPO_ROOT)/scripts/check-go-env.sh "$(REPO_ROOT)"

native: check-go-env copy-service
	go install $(REPO_ROOT)/tools/bbbsupervisor

aarch64: check-go-env copy-service
	GOARCH=arm64 go install $(REPO_ROOT)/tools/bbbsupervisor
	cp ${GOPATH}/bin/linux_arm64/bbbsupervisor $(REPO_ROOT)/build/

copy-service:
	cp bbbsupervisor.service $(REPO_ROOT)/build/
