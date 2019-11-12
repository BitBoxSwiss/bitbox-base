.DEFAULT_GOAL := aarch64
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/..
SHELL:=/bin/bash
RPCPORT:=18443
RPCUSER:=rpcuser
RPCPASSWORD:=rpcpass

check-go-env:
	@echo "Checking that environment supports Go builds.."
	@$(REPO_ROOT)/contrib/check-go-env.sh "$(REPO_ROOT)"

ci: generate
	cd $(REPO_ROOT)/middleware/src
	go test ./...
	golangci-lint run

envinit:
	@echo "Initializing Go environment.."
	$(REPO_ROOT)/middleware/contrib/envinit.sh

native: check-go-env ci
	go install $(REPO_ROOT)/middleware/cmd/middleware

build: generate
	go install $(REPO_ROOT)/middleware/cmd/middleware

aarch64: check-go-env ci
	GOARCH=arm64 go build $(REPO_ROOT)/middleware/cmd/middleware
	cp $(REPO_ROOT)/middleware/middleware $(REPO_ROOT)/bin/go/bbbmiddleware

regtest-up:
	cd $(REPO_ROOT)/middleware/integration_test ;\
	docker-compose up -d ;\
	sudo chown ${USER} volumes/clightning1/lightning-rpc ;\
	sudo chown ${USER} volumes/clightning2/lightning-rpc ;\
	docker-compose exec -d bitcoind sh /scripts/prometheus-bitcoind.sh ;\

regtest-down:
	cd $(REPO_ROOT)/middleware/integration_test ;\
	docker-compose down ;\

# Creates initial 101 blocks so we have spendable output
regtest-init:
	cd $(REPO_ROOT)/middleware/integration_test ;\
	docker-compose exec bitcoind sh /scripts/regtest-init.sh ;\

# Generates additional 5 new blocks
regtest-generate:
	cd $(REPO_ROOT)/middleware/integration_test ;\
	docker-compose exec bitcoind sh /scripts/generate5.sh ;\

regtest-info:
	cd $(REPO_ROOT)/middleware/integration_test ;\
	docker-compose exec bitcoind bitcoin-cli -regtest -rpcport=${RPCPORT} -rpcuser=${RPCUSER} -rpcpassword=${RPCPASSWORD} getblockchaininfo

generate:
	go generate ./...
