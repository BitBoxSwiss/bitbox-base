// Package middleware emits events with data from services running on the base.
package middleware

import (
	"log"
	"os/exec"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	lightning "github.com/fiatjaf/lightningd-gjson-rpc"
)

//go:generate protoc --go_out=import_path=messages:. messages/bbb.proto
const (
	opUCanHasDemo = "d"
)

// Middleware connects to services on the base with provided parrameters and emits events for the handler.
type Middleware struct {
	info           SampleInfoResponse
	environment    system.Environment
	events         chan []byte
	electrumEvents chan []byte
}

// NewMiddleware returns a new instance of the middleware
func NewMiddleware(argumentMap map[string]string) *Middleware {
	middleware := &Middleware{
		environment: system.NewEnvironment(argumentMap),
		//TODO(TheCharlatan) find a better way to increase the channel size
		events:         make(chan []byte), //the channel size needs to be increased every time we had an extra endpoint
		electrumEvents: make(chan []byte),
		info: SampleInfoResponse{
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

// demoCLightningRPC demonstrates a connection with lightnind. Currently it gets the lightningd alias and writes it into the SampleInfoResponse.
func (middleware *Middleware) demoCLightningRPC() {
	ln := &lightning.Client{
		Path: middleware.environment.GetLightningRPCPath(),
	}

	nodeinfo, err := ln.Call("getinfo")
	if err != nil {
		log.Println(err.Error() + " Lightningd getinfo called failed.")
		return
	}
	middleware.info.LightningAlias = nodeinfo.Get("alias").String()
}

//TODO rpcLoop just sends an event to the first client that catches it. In future, this information should properly fan out to all connected clients.
func (middleware *Middleware) rpcLoop() {
	for {
		middleware.demoBitcoinRPC()
		middleware.demoCLightningRPC()
		middleware.events <- []byte(opUCanHasDemo)
		time.Sleep(5 * time.Second)
	}
}

// Start gives a trigger for the handler to start the rpc event loop
func (middleware *Middleware) Start() <-chan []byte {
	go middleware.rpcLoop()
	return middleware.events
}

// ResyncBitcoinResponse is the struct that gets sent by the rpc server during a ResyncBitcoin call
type ResyncBitcoinResponse struct {
	Success bool
}

// GetEnvResponse is the struct that gets sent by the rpc server during a GetSystemEnv call
type GetEnvResponse struct {
	Network        string
	ElectrsRPCPort string
}

// SampleInfo holds sample information from c-lightning and bitcoind. It is temporary for testing purposes
type SampleInfoResponse struct {
	Blocks         int64   `json:"blocks"`
	Difficulty     float64 `json:"difficulty"`
	LightningAlias string  `json:"lightningAlias"`
}

// ResyncBitcoin returns a ResyncBitcoinResponse struct in response to a rpcserver request
func (middleware *Middleware) ResyncBitcoin() ResyncBitcoinResponse {
	cmd := exec.Command("."+middleware.environment.GetBBBConfigScript(), "exec", "bitcoin_reindex")
	err := cmd.Run()
	response := ResyncBitcoinResponse{Success: true}
	if err != nil {
		log.Println(err.Error() + " failed to run resync command, script does not exist")
		response = ResyncBitcoinResponse{Success: false}
	}
	return response
}

// SystemEnv returns a GetEnvResponse struct in response to a rpcserver request
func (middleware *Middleware) SystemEnv() GetEnvResponse {
	response := GetEnvResponse{
		Network:        middleware.environment.Network,
		ElectrsRPCPort: middleware.environment.ElectrsRPCPort,
	}
	return response
}

// SampleInfo returns a SampleInfoResponse struct in response to a rpcserver request
func (middleware *Middleware) SampleInfo() SampleInfoResponse {
	return middleware.info
}
