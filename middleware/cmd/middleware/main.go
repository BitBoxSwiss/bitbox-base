// Package main provides the entry point into the middleware and accepts command line arguments.
// Once compiled, the application pipes information from bitbox-base backend services to the bitbox-wallet-app and serves as an authenticator to the bitbox-base.
package main

import (
	"flag"
	"log"
	"net/http"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/handlers"
)

func main() {
	bitcoinRPCUser := flag.String("rpcuser", "rpcuser", "Bitcoin rpc user name")
	bitcoinRPCPassword := flag.String("rpcpassword", "rpcpassword", "Bitcoin rpc password")
	bitcoinRPCPort := flag.String("rpcport", "18332", "Bitcoin rpc port, localhost is assumed as an address")
	lightningRPCPath := flag.String("lightning-rpc-path", "/home/bitcoin/.lightning/lightning-rpc", "Path to the lightning rpc unix socket")
	electrsRPCPort := flag.String("electrsport", "51002", "Electrs rpc port")
	dataDir := flag.String("datadir", ".base", "Directory where middleware persistent data like noise keys is stored")
	network := flag.String("network", "testnet", "Indicate wether running bitcoin on testnet or mainnet")
	flag.Parse()

	logBeforeExit := func() {
		// Recover from all panics and log error before panicking again.
		if r := recover(); r != nil {
			// r is of type interface{}, just print its value
			log.Printf("%v, error detected, shutting down.", r)
			panic(r)
		}
	}
	defer logBeforeExit()
	middleware := middleware.NewMiddleware(*bitcoinRPCUser, *bitcoinRPCPassword, *bitcoinRPCPort, *lightningRPCPath, *electrsRPCPort, *network)
	log.Println("--------------- Started middleware --------------")

	handlers := handlers.NewHandlers(middleware, *dataDir)
	log.Println("Binding middleware api to port 8845")

	if err := http.ListenAndServe(":8845", handlers.Router); err != nil {
		log.Println(err.Error() + " Failed to listen for HTTP")
	}
}
