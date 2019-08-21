// Package middleware emits events with data from services running on the base.
package middleware

import (
	"log"
	"os/exec"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/digitalbitbox/bitbox-base/middleware/src/prometheus"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	lightning "github.com/fiatjaf/lightningd-gjson-rpc"
)

// Middleware connects to services on the base with provided parrameters and emits events for the handler.
type Middleware struct {
	info                 rpcmessages.SampleInfoResponse
	environment          system.Environment
	events               chan []byte
	prometheusClient     *prometheus.PromClient
	verificationProgress rpcmessages.VerificationProgressResponse
}

// NewMiddleware returns a new instance of the middleware
func NewMiddleware(argumentMap map[string]string) *Middleware {
	middleware := &Middleware{
		environment: system.NewEnvironment(argumentMap),
		//TODO(TheCharlatan) find a better way to increase the channel size
		events: make(chan []byte), //the channel size needs to be increased every time we had an extra endpoint
		info: rpcmessages.SampleInfoResponse{
			Blocks:         0,
			Difficulty:     0.0,
			LightningAlias: "disconnected",
		},
		verificationProgress: rpcmessages.VerificationProgressResponse{
			Blocks:               0,
			Headers:              0,
			VerificationProgress: 0.0,
		},
	}
	middleware.prometheusClient = prometheus.NewPromClient(middleware.environment.GetPrometheusURL())

	return middleware
}

// demoBitcoinRPC is a function that demonstrates a connection to bitcoind. Currently it gets the blockcount and difficulty and writes it into the SampleInfo. Once the demo is no longer needed, it should be removed
func (middleware *Middleware) GetSampleInfo() bool {
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
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Println(err.Error() + " No blockcount received")
		return false
	}

	blockChainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		log.Println(err.Error() + " GetBlockChainInfo rpc call failed")
		return false
	}

	ln := &lightning.Client{
		Path: middleware.environment.GetLightningRPCPath(),
	}

	nodeinfo, err := ln.Call("getinfo")
	if err != nil {
		log.Println(err.Error() + " Lightningd getinfo called failed.")
		return false
	}

	updateInfo := rpcmessages.SampleInfoResponse{
		Blocks:         blockCount,
		Difficulty:     blockChainInfo.Difficulty,
		LightningAlias: nodeinfo.Get("alias").String(),
	}
	if updateInfo != middleware.info {
		middleware.info = updateInfo
		return true
	}
	return false

}

func (middleware *Middleware) GetVerificationProgress() bool {
	updateVerificationProgress := rpcmessages.VerificationProgressResponse{
		Blocks:               middleware.prometheusClient.Blocks(),
		Headers:              middleware.prometheusClient.Headers(),
		VerificationProgress: middleware.prometheusClient.VerificationProgress(),
	}
	if updateVerificationProgress != middleware.verificationProgress {
		middleware.verificationProgress = updateVerificationProgress
		return true
	}
	return false
}

// rpcLoop gets new data from the various rpc connections of the middleware and emits events if new data is available
func (middleware *Middleware) rpcLoop() {
	for {
		if middleware.GetSampleInfo() {
			middleware.events <- []byte(rpcmessages.OpUCanHasSampleInfo)
		}
		if middleware.GetVerificationProgress() {
			middleware.events <- []byte(rpcmessages.OpUCanHasVerificationProgress)
		}
		time.Sleep(5 * time.Second)
	}
}

// Start gives a trigger for the handler to start the rpc event loop
func (middleware *Middleware) Start() <-chan []byte {
	go middleware.rpcLoop()
	return middleware.events
}

// ResyncBitcoin returns a ResyncBitcoinResponse struct in response to a rpcserver request
func (middleware *Middleware) ResyncBitcoin(option rpcmessages.ResyncBitcoinArgs) (rpcmessages.ResyncBitcoinResponse, error) {
	var cmd *exec.Cmd
	switch option {
	case rpcmessages.Resync:
		log.Println("executing full bitcoin resync in config script")
		cmd = exec.Command("."+middleware.environment.GetBBBConfigScript(), "exec", "bitcoin_resync")
	case rpcmessages.Reindex:
		log.Println("executing bitcoin reindex in config script")
		cmd = exec.Command("."+middleware.environment.GetBBBConfigScript(), "exec", "bitcoin_reindex")
	default:
	}
	err := cmd.Run()
	response := rpcmessages.ResyncBitcoinResponse{Success: true}
	if err != nil {
		log.Println(err.Error() + " failed to run resync command, script does not exist")
		response = rpcmessages.ResyncBitcoinResponse{Success: false}
	}
	return response, nil
}

// SystemEnv returns a new GetEnvResponse struct with the values as read from the environment
func (middleware *Middleware) SystemEnv() rpcmessages.GetEnvResponse {
	response := rpcmessages.GetEnvResponse{Network: middleware.environment.Network, ElectrsRPCPort: middleware.environment.ElectrsRPCPort}
	return response
}

// SampleInfo returns the chached SampleInfoResponse struct
func (middleware *Middleware) SampleInfo() rpcmessages.SampleInfoResponse {
	return middleware.info
}

// VerificationProgress returns the cached VerificationProgressResponse struct
func (middleware *Middleware) VerificationProgress() rpcmessages.VerificationProgressResponse {
	return middleware.verificationProgress
}

// Flashdrive returns a FlashdriveResponse struct in response to a rpcserver request
func (middleware *Middleware) Flashdrive(args rpcmessages.FlashdriveArgs) (rpcmessages.FlashdriveResponse, error) {
	switch args.Method {
	case rpcmessages.Check:
		log.Println("executing a USB flashdrive check via the cmd script")
		out, err := middleware.runBBBCmdScript("usb_flashdrive", "check")
		if err != nil {
			return rpcmessages.FlashdriveResponse{Success: false, Message: string(out)}, nil
		}
		return rpcmessages.FlashdriveResponse{Success: true, Message: string(out)}, nil

	case rpcmessages.Mount:
		log.Println("executing a USB flashdrive mount via the cmd script")
		out, err := middleware.runBBBCmdScript("usb_flashdrive", "mount"+" "+args.Path)
		if err != nil {
			return rpcmessages.FlashdriveResponse{Success: false, Message: string(out)}, nil
		}
		return rpcmessages.FlashdriveResponse{Success: true, Message: string(out)}, nil

	case rpcmessages.Unmount:
		log.Println("executing a USB flashdrive unmount via the cmd script")
		out, err := middleware.runBBBCmdScript("usb_flashdrive", "unmount")
		if err != nil {
			return rpcmessages.FlashdriveResponse{Success: false, Message: string(out)}, nil
		}
		return rpcmessages.FlashdriveResponse{Success: true, Message: string(out)}, nil

	default:
		return rpcmessages.FlashdriveResponse{Success: false, Message: "FlashdriveMethod not supported. (" + string(args.Method) + ")"}, nil
	}
}

func (middleware *Middleware) runBBBCmdScript(method string, arg string) (out []byte, err error) {
	script := middleware.environment.GetBBBCmdScript()
	cmdAsString := "." + script + " " + method + " " + arg
	out, err = exec.Command("."+script, method, "check").Output()
	if err != nil {
		// no error handling here, only logging.
		log.Printf("Error: The command '%s' exited with the output '%v' and error '%s'.\n", cmdAsString, string(out), err.Error())
	}
	return
}
