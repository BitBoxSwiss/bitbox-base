// Package middleware emits events with data from services running on the base.
package middleware

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/digitalbitbox/bitbox-base/middleware/src/prometheus"
	"github.com/digitalbitbox/bitbox-base/middleware/src/redis"
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
	redisClient          redis.Redis
	verificationProgress rpcmessages.VerificationProgressResponse
	// Saves state for the dummy setup process
	// TODO: should be removed as soon as Authentication is implemented
	dummyIsBaseSetup   bool
	dummyAdminPassword string
}

// GetMiddlewareVersion returns the Middleware Version for the `GET /version` endpoint.
func (middleware *Middleware) GetMiddlewareVersion() string {
	return middleware.environment.GetMiddlewareVersion()
}

// NewMiddleware returns a new instance of the middleware.
// For testing a mock boolean can be passed, which mocks e.g. redis.
func NewMiddleware(argumentMap map[string]string, mock bool) *Middleware {
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

	if !mock {
		middleware.redisClient = redis.NewClient(middleware.environment.GetRedisPort())
	} else if mock {
		middleware.redisClient = redis.NewMockClient("")
	}
	return middleware
}

// GetSampleInfo is a function that demonstrates a connection to bitcoind. Currently it gets the blockcount and difficulty and writes it into the SampleInfo. Once the demo is no longer needed, it should be removed
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
	log.Println("executing full bitcoin resync via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"bitcoind", "resync"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// ReindexBitcoin returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) ReindexBitcoin() rpcmessages.ErrorResponse {
	log.Println("executing full bitcoin resync via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"bitcoind", "reindex"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
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

// BackupSysconfig creates a backup of the system configuration onto a flashdrive.
// 1. Check if one and only one valid flashdrive is plugged in
// 2. Mount the flashdrive
// 3. Backup the system configuration
// 4. Unmount the flashdrive
func (middleware *Middleware) BackupSysconfig() (response rpcmessages.ErrorResponse) {

	response = middleware.mountFlashdrive()
	if !response.Success {
		return response
	}

	// It's crucial that mounted flashdrives get unmounted.
	defer func() {
		unmountResponse := middleware.unmountFlashdrive()
		// In case the backing up the system configuration fails the error message should
		// be preserved. If the backup was successful, but the unmouting fails, then the
		// ErrorCode and message should be overwritten.
		if response.Success {
			response = unmountResponse // overrites the backup response
		}
	}()

	log.Println("Executing a backup of the system config via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"backup", "sysconfig"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorBackupSysconfigNotAMountpoint,
		})

		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// BackupHSMSecret returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) BackupHSMSecret() rpcmessages.ErrorResponse {
	log.Println("Executing a backup of the c-lightning hsm_secret via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"backup", "hsm_secret"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// RestoreSysconfig restores a backup of the system configuration from the flashdrive.
// 1. Check if one and only one valid flashdrive is plugged in
// 2. Mount the flashdrive
// 3. Restore the system configuration (currently not choosable)
// 4. Unmount the flashdrive
func (middleware *Middleware) RestoreSysconfig() (response rpcmessages.ErrorResponse) {
	response = middleware.mountFlashdrive()
	if !response.Success {
		return response
	}

	// It's crucial that mounted flashdrives get unmounted.
	defer func() {
		unmountResponse := middleware.unmountFlashdrive()
		// In case the restoring up the system configuration fails the error message should
		// be preserved. If the backup was successful, but the unmouting fails, then the
		// ErrorCode and message should be overwritten.
		if response.Success {
			response = unmountResponse // overrites the backup response
		}
	}()

	log.Println("Executing a restore of the system config via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"restore", "sysconfig"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorRestoreSysconfigBackupNotFound,
		})

		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// RestoreHSMSecret returns a ErrorResponse struct in response to a rpcserver request
func (middleware *Middleware) RestoreHSMSecret() rpcmessages.ErrorResponse {
	log.Println("Executing a restore of the c-lightning hsm_secret via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"restore", "hsm_secret"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
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

	return rpcmessages.ErrorResponse{
		Success: false,
		Message: "authentication unsuccessful",
		Code:    rpcmessages.ErrorDummyAuthenticationNotSuccessful,
	}
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

	return rpcmessages.ErrorResponse{
		Success: false,
		Message: "password change unsuccessful (too short)",
		Code:    rpcmessages.ErrorDummyPasswordTooShort,
	}
}

// SetHostname sets the systems hostname
func (middleware *Middleware) SetHostname(args rpcmessages.SetHostnameArgs) rpcmessages.ErrorResponse {
	log.Println("Setting the hostname via the config script")
	var r = regexp.MustCompile(`^[a-z][a-z0-9-]{0,22}[a-z0-9]$`)
	hostname := args.Hostname

	if r.MatchString(hostname) {
		out, err := middleware.runBBBConfigScript([]string{"set", "hostname", hostname})
		if err != nil {
			errorCode := handleBBBScriptErrorCode(out, err, []rpcmessages.ErrorCode{
				rpcmessages.ErrorSetHostnameInvalidValue,
			})

			return rpcmessages.ErrorResponse{
				Success: false,
				Message: strings.Join(out, "\n"),
				Code:    errorCode,
			}
		}
		return rpcmessages.ErrorResponse{Success: true}
	}
	return rpcmessages.ErrorResponse{Success: false, Message: "invalid hostname"}
}

// GetHostname returns a the systems hostname in a GetHostnameResponse
func (middleware *Middleware) GetHostname() rpcmessages.GetHostnameResponse {
	// TODO: Implement get hostname from redis
	// TODO: define error codes for this
	return rpcmessages.GetHostnameResponse{
		ErrorResponse: &rpcmessages.ErrorResponse{
			Success: false,
			Message: "GetHostname is not implemnted",
			Code:    rpcmessages.ErrorUnexpected,
		},
	}
}

// EnableTor enables/disables the tor.service and configures bitcoind and lightningd based on the passed ToggleSettingEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTor(toggleAction rpcmessages.ToggleSetting) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor: %t' via the config script.\n", toggleAction)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "tor"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// EnableTorMiddleware enables/disables the tor hidden service for the middleware based on the passed ToggleSettingEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorMiddleware(toggleAction rpcmessages.ToggleSetting) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor for middleware: %t' via the config script.\n", toggleAction)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "tor_bbbmiddleware"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// EnableTorElectrs enables/disables the tor hidden service for electrs based on the passed ToggleSettingEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorElectrs(toggleAction rpcmessages.ToggleSetting) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor for electrs: %t' via the config script.\n", toggleAction)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "tor_electrs"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// EnableTorSSH enables/disables the tor hidden service for ssh based on the passed ToggleSettingEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorSSH(toggleAction rpcmessages.ToggleSetting) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor for ssh: %t' via the config script.\n", toggleAction)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "tor_ssh"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// EnableClearnetIBD enables/disables the initial block download over clearnet based on the passed ToggleSettingEnable/Disable argument
func (middleware *Middleware) EnableClearnetIBD(toggleAction rpcmessages.ToggleSetting) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable clearnet IBD: %t' via the config script.\n", toggleAction)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "bitcoin_ibd_clearnet"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorSetNeedsTwoArguments,
			rpcmessages.ErrorEnableClearnetIBDTorAlreadyDisabled,
		})

		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// ShutdownBase returns an ErrorResponse struct in response to a rpcserver request
// It calls the bbb-cmd.sh script which initializes a shutdown
func (middleware *Middleware) ShutdownBase() rpcmessages.ErrorResponse {
	log.Println("shutting down the Base via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"base", "shutdown"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// RebootBase returns an ErrorResponse struct in response to a rpcserver request
// It calls the bbb-cmd.sh script which initializes a restart
func (middleware *Middleware) RebootBase() rpcmessages.ErrorResponse {
	log.Println("restarting the Base via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"base", "restart"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// GetBaseVersion returns an GetBaseVersionResponse struct containing the base version in response to a rpcserver request
func (middleware *Middleware) GetBaseVersion() rpcmessages.GetBaseVersionResponse {
	log.Println("getting the Base version from redis")
	const baseVersionKey string = "base:version"
	version, err := middleware.redisClient.GetString(baseVersionKey)
	if err != nil {
		return rpcmessages.GetBaseVersionResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: fmt.Errorf("could not get %s from redis: %s", baseVersionKey, err.Error()).Error(),
				Code:    rpcmessages.ErrorRedisError,
			},
		}
	}
	return rpcmessages.GetBaseVersionResponse{
		ErrorResponse: &rpcmessages.ErrorResponse{
			Success: true,
		},
		Version: version,
	}
}

// EnableRootLogin enables/disables the login via the root user/password
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableRootLogin(toggleAction rpcmessages.ToggleSetting) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable root login: %t' via the config script.\n", toggleAction)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "root_pwlogin"})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, nil)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
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
		out, err := middleware.runBBBConfigScript([]string{"set", "root_pw", password})
		if err != nil {
			errorCode := handleBBBScriptErrorCode(out, err, []rpcmessages.ErrorCode{
				rpcmessages.ErrorSetNeedsTwoArguments,
			})

			return rpcmessages.ErrorResponse{
				Success: false,
				Message: strings.Join(out, "\n"),
				Code:    errorCode,
			}
		}

		return rpcmessages.ErrorResponse{Success: true}
	}

	return rpcmessages.ErrorResponse{
		Success: false,
		Message: "The password has to be at least 8 chars. An unicode char is counted as one.",
		Code:    rpcmessages.ErrorSetRootPasswordTooShort,
	}
}
