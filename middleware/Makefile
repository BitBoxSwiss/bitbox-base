.DEFAULT_GOAL := aarch64

ci:
	./scripts/ci.sh
envinit:
	./scripts/go-get.sh v1.15.0 github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u github.com/golang/dep/cmd/dep
native:
	go build -o ../build/base-middleware ./src/
aarch64:
	GOARCH=arm64 go build -o ../build/base-middleware ./src/