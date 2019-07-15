# BitBox Base Middleware

### THIS IS DEMO CODE AND NOT MEANT FOR PRODUCTION USE

This project serves as a communication hub between bitcoin core, electrum-x,
c-lightning and further services that run on the base and the
bitbox-wallet-app.

It's architecture should be able to handle multiple clients, should run a
websocket api and by default only expose its api to noise authenticated
clients.

## Developing

Currently, to build an run, install go and run:

    make native

for a native build. You can also cross compile to arm64/aarch64 with

    make aarch64

Before committing be sure to run `gofmt -w *` to properly indent the code.

You can also run `make envinit` to setup a development environment (dep and ci
tools)

If the protobuf messages need to be re-generated a specific version of protobuf
needs to be installed. The current messages are compiled with protobuf v1.2.0 .
To install a specific version locally, run:

```
GIT_TAG="v1.2.0" # change as needed
go get -d -u github.com/golang/protobuf/protoc-gen-go
git -C "$(go env GOPATH)"/src/github.com/golang/protobuf checkout $GIT_TAG
go install github.com/golang/protobuf/protoc-gen-go
```

## Running

The middleware accepts some command line arguments to connect to c-lightning
and bitcoind. The arguments are expected to be passed in the following fashion:

    ./base-middleware -rpcuser rpcuser -rpcpassword rpcpassword -rpcport 18332 -lightning-rpc-path /home/bitcoin/.lightning/lightning-rpc

Running `./base-middleware -h` will print the following help:
    
    -lightning-rpc-path string
    	Path to the lightning rpc unix socket (default "/home/bitcoin/.lightning/lightning-rpc")
    -rpcpassword string
      	Bitcoin rpc password (default "rpcpassword")
    -rpcport string
      	Bitcoin rpc port, localhost is assumed as an address (default "8332")
    -rpcuser string
    	Bitcoin rpc user name (default "rpcuser")


