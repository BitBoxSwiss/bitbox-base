// Package middleware emits events with data from services running on the base.
package middleware

import (
	"log"
	"os/exec"
	"regexp"
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

const (
	// Since enable and disable are two often-used parameters passed to bbb-config
	// and golangci-lint fails because of "string `disable` has 3 occurrences, make it a constant (goconst)"
	// they are constants.
	enableAction  string = "enable"
	disableAction string = "disable"
)

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

// ResyncBitcoin returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) ResyncBitcoin() rpcmessages.ErrorResponse {
	log.Println("executing full bitcoin resync via the config script")
	out, err := middleware.runBBBCmdScript("bitcoind", "resync", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// ReindexBitcoin returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) ReindexBitcoin() rpcmessages.ErrorResponse {
	log.Println("executing full bitcoin resync via the config script")
	out, err := middleware.runBBBCmdScript("bitcoind", "reindex", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
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

// MountFlashdrive returns an ErrorResponse struct in a response to a rpcserver request
func (middleware *Middleware) MountFlashdrive() rpcmessages.ErrorResponse {
	log.Println("Executing a USB flashdrive check via the cmd script")
	outCheck, err := middleware.runBBBCmdScript("flashdrive", "check", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(outCheck), Code: err.Error()}
	}
	flashDriveName := strings.TrimSuffix(string(outCheck), "\n")

	log.Println("Executing a USB flashdrive mount via the cmd script")
	outMount, err := middleware.runBBBCmdScript("flashdrive", "mount", flashDriveName)
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(outMount), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// UnmountFlashdrive returns an ErrorResponse struct in a response to a rpcserver request
func (middleware *Middleware) UnmountFlashdrive() rpcmessages.ErrorResponse {
	log.Println("Executing a USB flashdrive unmount via the cmd script")
	out, err := middleware.runBBBCmdScript("flashdrive", "unmount", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// BackupSysconfig returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) BackupSysconfig() rpcmessages.ErrorResponse {
	log.Println("Executing a backup of the system config via the cmd script")
	out, err := middleware.runBBBCmdScript("backup", "sysconfig", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// BackupHSMSecret returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) BackupHSMSecret() rpcmessages.ErrorResponse {
	log.Println("Executing a backup of the c-lightning hsm_secret via the cmd script")
	out, err := middleware.runBBBCmdScript("backup", "hsm_secret", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// RestoreSysconfig returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) RestoreSysconfig() rpcmessages.ErrorResponse {
	log.Println("Executing a restore of the system config via the cmd script")
	out, err := middleware.runBBBCmdScript("restore", "sysconfig", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// RestoreHSMSecret returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) RestoreHSMSecret() rpcmessages.ErrorResponse {
	log.Println("Executing a restore of the c-lightning hsm_secret via the cmd script")
	out, err := middleware.runBBBCmdScript("restore", "hsm_secret", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
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

// SetHostname sets the systems hostname
func (middleware *Middleware) SetHostname(args rpcmessages.SetHostnameArgs) rpcmessages.ErrorResponse {
	log.Println("Setting the hostname via the config script")
	var r = regexp.MustCompile(`^[a-z][a-z0-9-]{0,22}[a-z0-9]$`)
	hostname := args.Hostname

	if r.MatchString(hostname) {
		out, err := middleware.runBBBConfigScript("set", "hostname", hostname)
		if err != nil {
			return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
		}
		return rpcmessages.ErrorResponse{Success: true}
	}
	return rpcmessages.ErrorResponse{Success: false, Message: "invalid hostname"}
}

// GetHostname returns a the systems hostname in a GetHostnameResponse
func (middleware *Middleware) GetHostname() rpcmessages.GetHostnameResponse {
	log.Println("Getting the hostname via the config script")
	out, err := middleware.runBBBConfigScript("get", "hostname", "")
	if err != nil {
		return rpcmessages.GetHostnameResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: string(out),
				Code:    err.Error(),
			},
		}
	}

	hostname := strings.TrimSuffix(string(out), "\n")
	return rpcmessages.GetHostnameResponse{
		Hostname: hostname,
		ErrorResponse: &rpcmessages.ErrorResponse{
			Success: true,
		},
	}
}

// EnableTor enables/disables the tor.service and configures bitcoind and lightningd based on the passed boolean argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTor(enable bool) rpcmessages.ErrorResponse {
	var action string
	if enable {
		log.Println("Enabling Tor via the config script")
		action = enableAction
	} else {
		log.Println("Disabling Tor via the config script")
		action = disableAction
	}

	out, err := middleware.runBBBConfigScript(action, "tor", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// EnableTorMiddleware enables/disables the tor hidden service for the middleware based on the passed boolean argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorMiddleware(enable bool) rpcmessages.ErrorResponse {
	var action string
	if enable {
		log.Println("Enabling Tor for the middleware via the config script")
		action = enableAction
	} else {
		log.Println("Disabling Tor for the middleware via the config script")
		action = disableAction
	}

	out, err := middleware.runBBBConfigScript(action, "tor_bbbmiddleware", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// EnableTorElectrs enables/disables the tor hidden service for electrs based on the passed boolean argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorElectrs(enable bool) rpcmessages.ErrorResponse {
	var action string
	if enable {
		log.Println("Enabling Tor for electrs via the config script")
		action = enableAction
	} else {
		log.Println("Disabling Tor for electrs via the config script")
		action = disableAction
	}

	out, err := middleware.runBBBConfigScript(action, "tor_electrs", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// EnableTorSSH enables/disables the tor hidden service for ssh based on the passed boolean argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorSSH(enable bool) rpcmessages.ErrorResponse {
	var action string
	if enable {
		log.Println("Enabling Tor for ssh via the config script")
		action = enableAction
	} else {
		log.Println("Disabling Tor for ssh via the config script")
		action = disableAction
	}

	out, err := middleware.runBBBConfigScript(action, "tor_ssh", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// EnableClearnetIBD sets the initial block download over clearnet to either true or false
// based on the passed boolean argument and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableClearnetIBD(enable bool) rpcmessages.ErrorResponse {
	var value string
	if enable {
		log.Println("Setting clearnet IDB to true via the config script")
		value = "true"
	} else {
		log.Println("Setting clearnet IDB to false via the config script")
		value = "false"
	}

	out, err := middleware.runBBBConfigScript("set", "bitcoin_ibd_clearnet", value)
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// ShutdownBase returns an ErrorResponse struct in response to a rpcserver request
// It calls the bbb-cmd.sh script which initializes a shutdown
func (middleware *Middleware) ShutdownBase() rpcmessages.ErrorResponse {
	log.Println("shutting down the Base via the cmd script")
	out, err := middleware.runBBBCmdScript("base", "shutdown", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// RebootBase returns an ErrorResponse struct in response to a rpcserver request
// It calls the bbb-cmd.sh script which initializes a reboot
func (middleware *Middleware) RebootBase() rpcmessages.ErrorResponse {
	log.Println("rebooting the Base via the cmd script")
	out, err := middleware.runBBBCmdScript("base", "reboot", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// EnableRootLogin enables/disables the login via the root user/password
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableRootLogin(enable bool) rpcmessages.ErrorResponse {
	var action string
	if enable {
		log.Println("Enabling root login via the config script")
		action = enableAction
	} else {
		log.Println("Disabling root login via the config script")
		action = disableAction
	}

	out, err := middleware.runBBBConfigScript(action, "root_pwlogin", "")
	if err != nil {
		return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// SetRootPassword sets the systems root password
func (middleware *Middleware) SetRootPassword(args rpcmessages.SetRootPasswordArgs) rpcmessages.ErrorResponse {
	log.Println("Setting a new root password via the config script")
	password := args.RootPassword

	// Unicode passwords are allowed, but each Unicode rune is only counted as one when comparing the length
	// len("₿") = 3
	// len([]rune("₿")) = 1
	if len([]rune(password)) >= 8 {
		out, err := middleware.runBBBConfigScript("set", "root_pw", password)
		if err != nil {
			return rpcmessages.ErrorResponse{Success: false, Message: string(out), Code: err.Error()}
		}
		return rpcmessages.ErrorResponse{Success: true}
	}
	return rpcmessages.ErrorResponse{Success: false, Message: "invalid password"}
}

// runBBBCmdScript runs the bbb-cmd.sh script.
// The script executes commands like for example mounting a USB drive, doing a backup and copying files.
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

// runBBBConfigScript runs the bbb-config.sh script.
// The script changes the system configuration in redis by setting or unsetting the appropriate keys.
// If necessary the affected services are restarted.
func (middleware *Middleware) runBBBConfigScript(method string, arg1 string, arg2 string) (out []byte, err error) {
	script := middleware.environment.GetBBBConfigScript()
	cmdAsString := strings.Join([]string{script, method, arg1, arg2}, " ")
	out, err = exec.Command(script, method, arg1, arg2).Output()
	if err != nil {
		// no error handling here, only logging.
		log.Printf("Error: The command '%s' exited with the output '%v' and error '%s'.\n", cmdAsString, string(out), err.Error())
	}
	return
}
