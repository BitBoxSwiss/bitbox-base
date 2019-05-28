SHELL:=/bin/bash

envinit:
	./scripts/go-get.sh v1.15.0 github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u github.com/golang/dep/cmd/dep
native:
	cd src && go build -o base-middleware && mv base-middleware ..
aarch64:
	cd src && GOARCH=arm64 go build -o base-middleware && mv base-middleware ..
ci:
	./scripts/ci.sh

