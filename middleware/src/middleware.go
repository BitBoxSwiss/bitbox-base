// Package middleware emits events with data from services running on the base.
package middleware

import (
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	basemessages "github.com/digitalbitbox/bitbox-base/middleware/src/messages"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	lightning "github.com/fiatjaf/lightningd-gjson-rpc"

	"github.com/golang/protobuf/proto"
)

//go:generate protoc --go_out=import_path=messages:. messages/bbb.proto

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
	events      chan []byte
	mu          sync.RWMutex
}

// NewMiddleware returns a new instance of the middleware
func NewMiddleware(argumentMap map[string]string) *Middleware {
	middleware := &Middleware{
		environment: system.NewEnvironment(argumentMap),
		//TODO(TheCharlatan) find a better way to increase the channel size
		events: make(chan []byte), //the channel size needs to be increased every time we had an extra endpoint
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
		return
	}
	middleware.info.LightningAlias = nodeinfo.Get("alias").String()
}

//TODO rpcLoop just sends an event to the first client that catches it. In future, this information should properly fan out to all connected clients.
func (middleware *Middleware) rpcLoop() {
	for {
		middleware.demoBitcoinRPC()
		middleware.demoCLightningRPC()
		outgoing := &basemessages.BitBoxBaseOut{
			BitBoxBaseOut: &basemessages.BitBoxBaseOut_BaseMiddlewareInfoOut{
				BaseMiddlewareInfoOut: &basemessages.BaseMiddlewareInfoOut{
					Blocks:         middleware.info.Blocks,
					Difficulty:     float32(middleware.info.Difficulty),
					LightningAlias: middleware.info.LightningAlias,
				},
			},
		}
		response, err := proto.Marshal(outgoing)
		if err != nil {
			log.Println("Failed to marshal broadcast middlewareinfo outgoing message")
		}
		middleware.events <- response
		time.Sleep(5 * time.Second)
	}
}

// Start gives a trigger for the handler to start the rpc event loop
func (middleware *Middleware) Start() <-chan []byte {
	go middleware.rpcLoop()
	return middleware.events
}

// SystemEnv returns a protobuf serialized system environment information object
func (middleware *Middleware) SystemEnv() []byte {
	middleware.mu.Lock()
	defer middleware.mu.Unlock()
	outgoing := &basemessages.BitBoxBaseOut{
		BitBoxBaseOut: &basemessages.BitBoxBaseOut_BaseSystemEnvOut{
			BaseSystemEnvOut: &basemessages.BaseSystemEnvOut{
				Network:        middleware.environment.Network,
				ElectrsRPCPort: middleware.environment.ElectrsRPCPort,
			},
		},
	}
	response, err := proto.Marshal(outgoing)
	if err != nil {
		log.Println("Protobuf failed to marshal system env outgoing message")
	}
	return response
}

// Run resync command
func (middleware *Middleware) ResyncBitcoin() []byte {
	cmd := exec.Command("."+middleware.environment.GetBBBConfigScript(), "exec", "bitcoin_reindex")
	err := cmd.Run()
	if err != nil {
		log.Println(err.Error() + " failed to run resync command, script does not exist")
	}
	outgoing := &basemessages.BitBoxBaseOut{
		BitBoxBaseOut: &basemessages.BitBoxBaseOut_BaseResyncOut{
			BaseResyncOut: &basemessages.BaseResyncOut{},
		},
	}
	response, err := proto.Marshal(outgoing)
	if err != nil {
		log.Println("protobuf failed to marshal resyncBitcoin outgoing message")
	}
	return response
}
