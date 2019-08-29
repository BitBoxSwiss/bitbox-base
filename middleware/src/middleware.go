// Package middleware emits events with data from services running on the base.
package middleware

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
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
	// Saves state for the dummy setup process
	// TODO: should be removed as soon as Authentication is implemented
	dummyIsBaseSetup   bool
	dummyAdminPassword string
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
		dummyIsBaseSetup:   false,
		dummyAdminPassword: "",
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
		cmd = exec.Command(middleware.environment.GetBBBConfigScript(), "exec", "bitcoin_resync")
	case rpcmessages.Reindex:
		log.Println("executing bitcoin reindex in config script")
		cmd = exec.Command(middleware.environment.GetBBBConfigScript(), "exec", "bitcoin_reindex")
	default:
	}
	err := cmd.Run()
	if err != nil {
		log.Println(err.Error() + " failed to run resync command, script does not exist")
		response := rpcmessages.ResyncBitcoinResponse{Success: false}
		return response, err
	}
	response := rpcmessages.ResyncBitcoinResponse{Success: true}
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

// DummyIsBaseSetup returns the current dummyIsBaseSetup bool
// FIXME: this is a dummy function and should be removed once authentication is implemented
func (middleware *Middleware) DummyIsBaseSetup() bool {
	return middleware.dummyIsBaseSetup
}

// DummyAdminPassword returns the current dummyAdminPassword string
// FIXME: this is a dummy function and should be removed once authentication is implemented
func (middleware *Middleware) DummyAdminPassword() string {
	return middleware.dummyAdminPassword
}

// Flashdrive returns a GenericResponse struct in response to a rpcserver request
func (middleware *Middleware) Flashdrive(args rpcmessages.FlashdriveArgs) (rpcmessages.GenericResponse, error) {
	switch args.Method {
	case rpcmessages.Check:
		log.Println("Executing a USB flashdrive check via the cmd script")
		out, err := middleware.runBBBCmdScript("flashdrive", "check", "")
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	case rpcmessages.Mount:
		log.Println("Executing a USB flashdrive mount via the cmd script")
		out, err := middleware.runBBBCmdScript("flashdrive", "mount", args.Path)
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	case rpcmessages.Unmount:
		log.Println("Executing a USB flashdrive unmount via the cmd script")
		out, err := middleware.runBBBCmdScript("flashdrive", "unmount", "")
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	default:
		errorMessage := fmt.Sprintf("Method %d not supported for Flashdrive().", args.Method)
		return rpcmessages.GenericResponse{Success: false, Message: errorMessage}, errors.New(errorMessage)
	}
}

// Backup returns a GenericResponse struct in response to a rpcserver request
func (middleware *Middleware) Backup(method rpcmessages.BackupArgs) (rpcmessages.GenericResponse, error) {
	switch method {
	case rpcmessages.BackupSysConfig:
		log.Println("Executing a backup of the system config via the cmd script")
		out, err := middleware.runBBBCmdScript("backup", "sysconfig", "")
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	case rpcmessages.BackupHSMSecret:
		log.Println("Executing a backup of the c-lightning hsm_secret via the cmd script")
		out, err := middleware.runBBBCmdScript("backup", "hsm_secret", "")
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	default:
		errorMessage := fmt.Sprintf("Method %d not supported for Backup().", method)
		return rpcmessages.GenericResponse{Success: false, Message: errorMessage}, errors.New(errorMessage)
	}
}

// Restore returns a GenericResponse struct in response to a rpcserver request
func (middleware *Middleware) Restore(method rpcmessages.RestoreArgs) (rpcmessages.GenericResponse, error) {
	switch method {
	case rpcmessages.RestoreSysConfig:
		log.Println("Executing a restore of the system config via the cmd script")
		out, err := middleware.runBBBCmdScript("restore", "sysconfig", "")
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	case rpcmessages.RestoreHSMSecret:
		log.Println("Executing a restore of the c-lightning hsm_secret via the cmd script")
		out, err := middleware.runBBBCmdScript("restore", "hsm_secret", "")
		if err != nil {
			return rpcmessages.GenericResponse{Success: false, Message: string(out)}, err
		}
		return rpcmessages.GenericResponse{Success: true, Message: string(out)}, nil

	default:
		errorMessage := fmt.Sprintf("Method %d not supported for Restore().", method)
		return rpcmessages.GenericResponse{Success: false, Message: errorMessage}, errors.New(errorMessage)
	}
}

// UserAuthenticate returns an ErrorResponse struct in response to a rpcserver request.
// FIXME: This is a dummy implementation of the UserAuthenticate RPC call.
// in the future this should return an AuthentificationResponse with e.g. an JWT.
// This currently uses the `dummyIsBaseSetup` boolean, which should be removed when the proper authentication is implemented.
func (middleware *Middleware) UserAuthenticate(args rpcmessages.UserAuthenticateArgs) rpcmessages.ErrorResponse {

	// TODO: replace the dummyIsBaseSetup with a proper variable loaded from e.g. redis
	// dummyIsBaseSetup should only be used for the dummy UserAuthenticate RPC and gets reset on middleware restart
	if !middleware.dummyIsBaseSetup {
		if args.Username == "admin" && args.Password == "ICanHasPassword?" {
			// middleware.dummyIsBaseSetup is only set to true after the initial admin password is changed
			return rpcmessages.ErrorResponse{Success: true}
		}
	} else if middleware.dummyIsBaseSetup {
		dummyUsers := map[string]string{
			"admin":   middleware.dummyAdminPassword,
			"satoshi": "shift1",
			"dev":     "dev",
		}
		if expectedPasssword, userIsInMap := dummyUsers[args.Username]; userIsInMap {
			if args.Password == expectedPasssword {
				return rpcmessages.ErrorResponse{Success: true}
			}
		}
	}

	return rpcmessages.ErrorResponse{Success: false, Message: "authentication unsuccessful"}
}

// UserChangePassword returns an ErrorResponse struct in response to a rpcserver request
// FIXME: This is a dummy implementation of the UserChangePassword RPC call
// This dummy method approves all passwords which are longer or equal to 8 chars
func (middleware *Middleware) UserChangePassword(args rpcmessages.UserChangePasswordArgs) rpcmessages.ErrorResponse {

	if len(args.NewPassword) >= 8 {
		if args.Username == "admin" {
			middleware.dummyAdminPassword = args.NewPassword
			if !middleware.dummyIsBaseSetup {
				middleware.dummyIsBaseSetup = true // the change of the admin password completes the dummy setup process (for now)
			}
		}
		return rpcmessages.ErrorResponse{Success: true}
	}

	return rpcmessages.ErrorResponse{Success: false, Message: "password change unsuccessful (too short)"}
}

func (middleware *Middleware) runBBBCmdScript(method string, arg1 string, arg2 string) (out []byte, err error) {
	script := middleware.environment.GetBBBCmdScript()
	cmdAsString := strings.Join([]string{script, method, arg1, arg2}, " ")
	out, err = exec.Command(script, method, arg1, arg2).Output()
	if err != nil {
		// no error handling here, only logging.
		log.Printf("Error: The command '%s' exited with the output '%v' and error '%s'.\n", cmdAsString, string(out), err.Error())
	}
	return
}
