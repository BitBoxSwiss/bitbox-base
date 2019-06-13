// Package middleware emits events with data from services running on the base.
package middleware

import (
	"log"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	lightning "github.com/fiatjaf/lightningd-gjson-rpc"
)

// SampleInfo holds sample information from c-lightning and bitcoind. It is temporary for testing purposes.
type SampleInfo struct {
	Blocks         int64   `json:"blocks"`
	Difficulty     float64 `json:"difficulty"`
	LightningAlias string  `json:"lightningAlias"`
}

// Middleware connects to services on the base with provided parrameters and emits events for the handler.
type Middleware struct {
	info        SampleInfo
	environment system.Environment
	events      chan interface{}
}

// NewMiddleware returns a new instance of the middleware
func NewMiddleware(bitcoinRPCUser, bitcoinRPCPassword, bitcoinRPCPort, lightningRPCPath, electrsRPCPort, network string) *Middleware {
	middleware := &Middleware{
		environment: system.NewEnvironment(bitcoinRPCUser, bitcoinRPCPassword, bitcoinRPCPort, lightningRPCPath, electrsRPCPort, network),
		events:      make(chan interface{}),
		info: SampleInfo{
			Blocks:         0,
			Difficulty:     0.0,
			LightningAlias: "disconnected",
		},
	}

	return middleware
}

// demoBitcoinRPC is a function that demonstrates a connection to bitcoind. Currently it gets the blockcount and difficulty and writes it into the SampleInfo.
func (middleware *Middleware) demoBitcoinRPC() {
	connCfg := rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         "127.0.0.1:" + middleware.environment.GetBitcoinRPCPort(),
		User:         middleware.environment.GetBitcoinRPCUser(),
		Pass:         middleware.environment.GetBitcoinRPCPassword(),
	}
	client, err := rpcclient.New(&connCfg, nil)
	if err != nil {
		log.Println(err.Error() + " Failed to create new bitcoind rpc client")
	}
	//client is shutdown/deconstructed again as soon as this function returns
	defer client.Shutdown()

	//Get current block count.
	var blockCount int64
	blockCount, err = client.GetBlockCount()
	if err != nil {
		log.Println(err.Error() + " No blockcount received")
	} else {
		middleware.info.Blocks = blockCount
	}
	blockChainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		log.Println(err.Error() + " GetBlockChainInfo rpc call failed")
	} else {
		middleware.info.Difficulty = blockChainInfo.Difficulty
	}

}

// demoCLightningRPC demonstrates a connection with lightnind. Currently it gets the lightningd alias and writes it into the SampleInfo.
func (middleware *Middleware) demoCLightningRPC() {
	ln := &lightning.Client{
		Path: middleware.environment.GetLightningRPCPath(),
	}

	nodeinfo, err := ln.Call("getinfo")
	if err != nil {
		log.Println(err.Error() + " Lightningd getinfo called failed.")
	} else {
		middleware.info.LightningAlias = nodeinfo.Get("alias").String()
	}
}

func (middleware *Middleware) rpcLoop() {
	for {
		middleware.demoBitcoinRPC()
		middleware.demoCLightningRPC()
		middleware.events <- &middleware.info
		time.Sleep(5 * time.Second)
	}
}

// Start gives a trigger for the handler to start the rpc event loop
func (middleware *Middleware) Start() <-chan interface{} {
	go middleware.rpcLoop()
	return middleware.events
}

// GetSystemEnv implements a getter for the system environment.
func (middleware *Middleware) GetSystemEnv() system.Environment {
	return middleware.environment
}
