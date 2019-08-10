# BitBox Base Middleware

This project serves as a communication hub between bitcoin core, electrum-x,
c-lightning and further services that run on the base and the
bitbox-wallet-app.

## Overview

The middleware is able to handle multiple connected clients, runs a
websocket api and by default only exposes its api to noise authenticated
clients.

A connecting client can always probe availability on a single http endpoint. If
the middleware is available, it then opens a websocket connection and starts
a [noise 'XX' handshake](https://noiseprotocol.org/noise.html#handshake-patterns).
Once the handshake is complete the client's noise static public key is stored. If
this key is not in storage yet, no further communication with the middleware is
allowed, until the user manually verifies the [channel binding
hash](https://noiseprotocol.org/noise.html#channel-binding).

Once the user verfied the channel binding successfully, the middleware starts a
rpc server. Subsequent communication with the wallet app over the websocket is
then noise encrypted.  Each client connects to its own rpc server that in turn
executes function calls to a single middleware backend instance that manages and
controls data flow to and from bitcoind, lightningd and prometheus. To limit
calls to bitcoind during initial blockchain download and reindexing operations,
sync progress and other informational data is fetched from prometheus.

Data fetched from other services is cached in the middleware. These caching
data structs should be initialized when the middleware is instantiated. A connected
client should not trigger further rpc or http requests.

The message format of the rpc server is defined in the
[rpcmessages.go](src/rpcmessages/rpcmessages.go) file. To notify the wallet app
that new data in the app is available the middleware backend package emits
events unique to each rpc method that get transmitted as short encrypted websocket
messages to the wallet app. The wallet app can then call the respective rpc
methods.

## Developing

Currently, to build and run, install go and run:

    make native

for a native build. You can also cross compile to arm64/aarch64 with

    make aarch64

Before committing be sure to run `gofmt -w *` to properly indent the code.

You can also run `make envinit` to setup a development environment (dep and ci
tools)

## Running

The middleware accepts some command line arguments to connect to c-lightning
and bitcoind. The arguments are expected to be passed in the following fashion:

    ./bbbmiddleware -rpcuser rpcuser -rpcpassword rpcpassword -rpcport 18332 -lightning-rpc-path /home/bitcoin/.lightning/lightning-rpc

Running `./bbbmiddleware -h` will print the following help:
    
  -bbbconfigscript string
    	Path to the bbb-config file that allows setting system configuration (default "/opt/shift/scripts/bbb-config.sh")
  -datadir string
    	Directory where middleware persistent data like noise keys is stored (default ".base")
  -electrsport string
    	Electrs rpc port (default "51002")
  -lightning-rpc-path string
    	Path to the lightning rpc unix socket (default "/home/bitcoin/.lightning/lightning-rpc")
  -network string
    	Indicate wether running bitcoin on regtest, testnet or mainnet (default "testnet")
  -rpcpassword string
    	Bitcoin rpc password (default "rpcpassword")
  -rpcport string
    	Bitcoin rpc port, localhost is assumed as an address (default "18332")
  -rpcuser string
    	Bitcoin rpc user name (default "rpcuser")

## Testing

The Makefile also provides a target to run bitcoind, electrs and lightningd on
regtest in a docker container. Install docker-compose on your machine and run
`make regtest-up` to start the regtest setup and `make regtest-down` to shutdown.
The ports available are:

 - 18443 for bitcoin-cli
 - 60401 for electrs rpc

The current regtest docker-compose file will start two lightnind instances. Their
rpc files can be accessed in `integration_test/volumes/clightning1` and 
`intergation_test/volumes/clightning2` respectively.
To acces the lightningd unix port, the makefile target will ask for a sudo password 
to change permissions of the lightning-rpc file.

Once the docker container is up, use `bitcoin-cli` and `lightning-cli` to communicate:

    bitcoin-cli -regtest -rpcport=18443 -rpcuser=rpcuser -rpcpassword=rpcpass getblockchaininfo
    lightning-cli --rpc-file integration_test/volumes/clightning1/lightning-rpc getinfo

For the middleware run:

    middleware -rpcport=18443 -rpcpassword=rpcpass -rpcuser=rpcuser -electrsport=60401 -lightning-rpc-path=integration_test/volumes/clightning1/lightning-rpc

The two c-lightning instances allow communication between each other. To
connect clightning1 with clightning2, run getinfo on clightning2 and then
connect to its id on clightning1 with:

    docker exec -ti lightningd1 lightning-cli connect 026c213484d4b3cb8aff9d4186439bf4032b793831051cf1b4189c7d83a6ec47f1 10.10.0.13:9735

