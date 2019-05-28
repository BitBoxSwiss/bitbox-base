.DEFAULT_GOAL := aarch64

native: copy-service
	go build -o ../../build/bbbfancontrol
aarch64: copy-service
	GOARCH=arm64 go build -o ../../build/bbbfancontrol
copy-service:
	cp bbbfancontrol.service ../../build/
