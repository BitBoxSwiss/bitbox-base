.DEFAULT_GOAL := aarch64
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..

native: copy-service
	go build -o $(REPO_ROOT)/build/bbbfancontrol
aarch64: copy-service
	GOARCH=arm64 go build -o $(REPO_ROOT)/build/bbbfancontrol
copy-service:
	cp bbbfancontrol.service $(REPO_ROOT)/build/
