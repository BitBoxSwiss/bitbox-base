// Package middleware emits events with data from services running on the base.
package middleware

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/digitalbitbox/bitbox-base/middleware/src/authentication"
	"github.com/digitalbitbox/bitbox-base/middleware/src/prometheus"
	"github.com/digitalbitbox/bitbox-base/middleware/src/redis"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	lightning "github.com/fiatjaf/lightningd-gjson-rpc"
	"golang.org/x/crypto/bcrypt"
)

// UserAuthStruct holds the structure that is written into the redis middleware:auth key's value.
type UserAuthStruct struct {
	BCryptedPassword string `json:"password"`
	Role             string `json:"role"`
}

// initialAdminPassword is the default password that allows login when setting up a base.
const initialAdminPassword = "ICanHasPasword?"

// Middleware connects to services on the base with provided parameters and emits events for the handler.
type Middleware struct {
	info                 rpcmessages.SampleInfoResponse
	environment          system.Environment
	events               chan []byte
	prometheusClient     prometheus.Client
	redisClient          redis.Redis
	jwtAuth              *authentication.JwtAuth
	verificationProgress rpcmessages.VerificationProgressResponse
	serviceInfo          rpcmessages.GetServiceInfoResponse
	baseUpdateProgress   rpcmessages.GetBaseUpdateProgressResponse
	isMockmode           bool
	// Saves state for the setup process
	isMiddlewarePasswordSet bool
	isBaseSetupDone         bool
}

// GetMiddlewareVersion returns the Middleware Version for the `GET /version` endpoint.
func (middleware *Middleware) GetMiddlewareVersion() string {
	return middleware.environment.GetMiddlewareVersion()
}

// NewMiddleware returns a new instance of the middleware.
// For testing a mock boolean can be passed, which mocks e.g. redis.
func NewMiddleware(argumentMap map[string]string, mock bool) (*Middleware, error) {
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
		serviceInfo: rpcmessages.GetServiceInfoResponse{},
		baseUpdateProgress: rpcmessages.GetBaseUpdateProgressResponse{
			State:                 rpcmessages.UpdateNotInProgress,
			ProgressPercentage:    0,
			ProgressDownloadedKiB: 0,
		},
		isMockmode:              mock,
		isMiddlewarePasswordSet: false,
	}
	middleware.prometheusClient = prometheus.NewClient(middleware.environment.GetPrometheusURL())

	if !mock {
		middleware.redisClient = redis.NewClient(middleware.environment.GetRedisPort())
	} else if mock {
		middleware.redisClient = redis.NewMockClient("")
	}
	err := middleware.checkMiddlewareSetup()
	if err != nil {
		log.Println("failed to update the middleware password set flag")
		return &Middleware{}, err
	}
	if !middleware.isMiddlewarePasswordSet {
		usersMap := make(map[string]UserAuthStruct)
		bcryptedPassword, err := bcrypt.GenerateFromPassword([]byte(initialAdminPassword), 12)
		if err != nil {
			log.Println("Failed to generate new standard password")
			return &Middleware{}, err
		}
		usersMap["admin"] = UserAuthStruct{BCryptedPassword: string(bcryptedPassword), Role: "admin"}

		authStructureString, err := json.Marshal(&usersMap)
		if err != nil {
			log.Println("Unable to marshal auth structure map")
			return &Middleware{}, err
		}
		err = middleware.redisClient.SetString(redis.MiddlewareAuth, string(authStructureString))
		if err != nil {
			log.Println("Unable to initialize auth data structure")
			return &Middleware{}, err
		}
	}

	middleware.jwtAuth, err = authentication.NewJwtAuth()
	// return if there is an error, this should not really happen though on our device and in our dev environments, low entropy is usually common in embedded environments
	if err != nil {
		return &Middleware{}, err
	}

	return middleware, nil
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

// GetVerificationProgress makes calls to prometheus to get the current block, header and verification progress number and stores it in
// the middleware's verificationProgress struct. If the new data was available, the function returns true, if the data remained static, false.
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
		if middleware.didServiceInfoChange() {
			middleware.events <- []byte(rpcmessages.OpServiceInfoChanged)
		}
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

// SampleInfo returns the cached SampleInfoResponse struct
func (middleware *Middleware) SampleInfo() rpcmessages.SampleInfoResponse {
	return middleware.info
}

// VerificationProgress returns the cached VerificationProgressResponse struct
func (middleware *Middleware) VerificationProgress() rpcmessages.VerificationProgressResponse {
	return middleware.verificationProgress
}

// SetupStatus returns the current status in the setup process as a SetupStatusResponse struct. This includes the middleware password set boolean and the base setup boolean.
func (middleware *Middleware) SetupStatus() rpcmessages.SetupStatusResponse {
	return rpcmessages.SetupStatusResponse{MiddlewarePasswordSet: middleware.isMiddlewarePasswordSet, BaseSetup: middleware.isBaseSetupDone}
}

// InitialAdminPassword is a getter that returns the constant initialAdminPassword string
// This password is only valid for authentication until the admin user changes it.
func (middleware *Middleware) InitialAdminPassword() string {
	return initialAdminPassword
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

	//BackupSysconfig is the last call in the base setup. Once it is completed successfully, the setup is done.
	err = middleware.redisClient.SetString(redis.BaseSetupDone, "1")
	if err != nil {
		log.Println("redis failed to set the base setup done flag. This is not a critical error, but the setup will be repeated again")
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
// To check if the user should be authenticated from default values, the bool 'isMiddlewarePasswordSet' is read from redis
func (middleware *Middleware) UserAuthenticate(args rpcmessages.UserAuthenticateArgs) rpcmessages.UserAuthenticateResponse {
	// isMiddlewarePasswordSet checks if the base is run the first time.
	err := middleware.checkMiddlewareSetup()
	if err != nil {
		return rpcmessages.UserAuthenticateResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: "authentication failed, redis error",
				Code:    rpcmessages.ErrorAuthenticationFailed,
			},
		}
	}

	usersMap, err := middleware.getAuthStructure()
	if err != nil {
		return rpcmessages.UserAuthenticateResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: "authentication unsuccessful, see middleware logs for more information",
				Code:    rpcmessages.ErrorAuthenticationFailed,
			},
		}
	}

	if _, ok := usersMap[args.Username]; !ok {
		log.Printf("User %s not found in database", args.Username)
		//TODO: Once we support multiple users work over the ErrorAuthenticationUsernameNotFound ErrorResponse message. It reveals information about the database.
		return rpcmessages.UserAuthenticateResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: "authentication unsuccessful, username not found",
				Code:    rpcmessages.ErrorAuthenticationUsernameNotFound,
			},
		}
	}

	passwordFromStorage := usersMap[args.Username].BCryptedPassword
	err = bcrypt.CompareHashAndPassword([]byte(passwordFromStorage), []byte(args.Password))
	if err != nil {
		log.Println("Hash and password did not match")
		return rpcmessages.UserAuthenticateResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: "authentication unsuccessful, incorrect password",
				Code:    rpcmessages.ErrorAuthenticationPasswordIncorrect,
			},
		}
	}

	jwtTokenStr, err := middleware.jwtAuth.GenerateToken(args.Username)
	if err != nil {
		return rpcmessages.UserAuthenticateResponse{
			ErrorResponse: &rpcmessages.ErrorResponse{
				Success: false,
				Message: "authentication unsuccessful, jwt error",
				Code:    rpcmessages.ErrorAuthenticationFailed,
			},
		}
	}

	return rpcmessages.UserAuthenticateResponse{
		ErrorResponse: &rpcmessages.ErrorResponse{
			Success: true,
		},
		Token: jwtTokenStr,
	}
}

// UserChangePassword returns an ErrorResponse struct in response to a rpcserver request
// The function first validates the current password with redis, then replaces it with the new password.
// Passwords need to be longer than or equal to 8 chars.
func (middleware *Middleware) UserChangePassword(args rpcmessages.UserChangePasswordArgs) rpcmessages.ErrorResponse {
	if len(args.NewPassword) < 8 {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "password change unsuccessful, the password needs to be at least 8 characters in length",
			Code:    rpcmessages.ErrorPasswordTooShort,
		}
	}

	usersMap, err := middleware.getAuthStructure()
	if err != nil {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "authentication unsuccessful, see middleware logs for more information",
			Code:    rpcmessages.ErrorAuthenticationFailed,
		}
	}

	//TODO: Once we support multiple users work over the ErrorPasswordChangeUsernameNotExist message. It reveals information about the database.
	if _, ok := usersMap[args.Username]; !ok {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "username does not exist",
			Code:    rpcmessages.ErrorPasswordChangeUsernameNotExist,
		}
	}

	passwordFromStorage := usersMap[args.Username].BCryptedPassword
	err = bcrypt.CompareHashAndPassword([]byte(passwordFromStorage), []byte(args.Password))
	if err != nil {
		log.Println("Hash and password did not match")
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "password change unsuccessful, current password was incorrect",
			Code:    rpcmessages.ErrorPasswordChangePasswordIncorrect,
		}
	}

	bcryptedPassword, err := bcrypt.GenerateFromPassword([]byte(args.NewPassword), 12)
	if err != nil {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "password change unsuccessful",
			Code:    rpcmessages.ErrorPasswordChangeFailed,
		}
	}

	userAuthSecrets := usersMap[args.Username]
	userAuthSecrets.BCryptedPassword = string(bcryptedPassword)
	userAuthSecrets.Role = "admin"
	usersMap[args.Username] = userAuthSecrets
	usersMapByteStr, err := json.Marshal(usersMap)
	if err != nil {
		log.Println("Failed marshaling the new user data for redis")
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "password change unsuccessful",
			Code:    rpcmessages.ErrorPasswordChangeFailed,
		}
	}
	err = middleware.redisClient.SetString(redis.MiddlewareAuth, string(usersMapByteStr))
	if err != nil {
		log.Println("Failed committing the new password to redis")
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "password change unsuccessful",
			Code:    rpcmessages.ErrorPasswordChangeFailed,
		}
	}

	if !middleware.isMiddlewarePasswordSet {
		err := middleware.redisClient.SetString(redis.MiddlewarePasswordSet, "1")
		if err != nil {
			log.Println("Failed setting middleware password set to true")
		}
		middleware.isMiddlewarePasswordSet = true // the change of the admin password completes the setup process (for now)
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// ValidateToken validates a jwt token string and returns an error if not valid and nil otherwise.
func (middleware *Middleware) ValidateToken(token string) error {
	return middleware.jwtAuth.ValidateToken(token)
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

// EnableTor enables/disables the tor.service and configures bitcoind and lightningd based on the passed ToggleSettingArgsEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTor(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor: %t' via the config script.\n", toggleAction.ToggleSetting)
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

// EnableTorMiddleware enables/disables the tor hidden service for the middleware based on the passed ToggleSettingArgsEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorMiddleware(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor for middleware: %t' via the config script.\n", toggleAction.ToggleSetting)
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

// EnableTorElectrs enables/disables the tor hidden service for electrs based on the passed ToggleSettingArgsEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorElectrs(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor for electrs: %t' via the config script.\n", toggleAction.ToggleSetting)
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

// EnableTorSSH enables/disables the tor hidden service for ssh based on the passed ToggleSettingArgsEnable/Disable argument
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableTorSSH(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable Tor for ssh: %t' via the config script.\n", toggleAction.ToggleSetting)
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

// EnableClearnetIBD enables/disables the initial block download over clearnet based on the passed ToggleSettingArgsEnable/Disable argument
func (middleware *Middleware) EnableClearnetIBD(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable clearnet IBD: %t' via the config script.\n", toggleAction.ToggleSetting)
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

// ShutdownBase shuts the Base down.
// The shutdown is executed in a goroutine with a delay of a few seconds.
// Prior to starting the goroutine the path for the `shutdown` executable is checked.
// If the executable is found, a ErrorResponse indicating success is returned.
// Otherwise a ExecutableNotFound Code is returned.
func (middleware *Middleware) ShutdownBase() rpcmessages.ErrorResponse {
	const shutdownDelay time.Duration = 5 * time.Second
	log.Printf("Shutting down the Base in %s\n", shutdownDelay)

	if middleware.isMockmode {
		return rpcmessages.ErrorResponse{Success: true}
	}

	_, err := exec.LookPath("shutdown")
	if err != nil {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("could not shut the Base down: %s", err.Error()),
			Code:    rpcmessages.ExecutableNotFound,
		}
	}

	go func(delay time.Duration) {
		time.Sleep(delay)
		cmd := exec.Command("shutdown", "now")
		err = cmd.Start()
		if err != nil {
			log.Printf("Could not shutdown the Base: %s", err.Error())
		}
	}(shutdownDelay)

	return rpcmessages.ErrorResponse{Success: true}
}

// RebootBase reboots the Base.
// The reboot is executed in a goroutine with a delay of a few seconds.
// Prior to starting the goroutine the path for the `reboot` executable is checked.
// If the executable is found, a ErrorResponse indicating success is returned.
// Otherwise a ExecutableNotFound Code is returned.
func (middleware *Middleware) RebootBase() rpcmessages.ErrorResponse {
	const rebootDelay time.Duration = 5 * time.Second
	log.Printf("Rebooting the Base in %s\n", rebootDelay)

	if middleware.isMockmode {
		return rpcmessages.ErrorResponse{Success: true}
	}

	_, err := exec.LookPath("reboot")
	if err != nil {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("could not reboot the Base: %s", err.Error()),
			Code:    rpcmessages.ExecutableNotFound,
		}
	}

	go func(delay time.Duration) {
		time.Sleep(delay)
		cmd := exec.Command("reboot")
		err = cmd.Start()
		if err != nil {
			log.Printf("Could not reboot the Base: %s", err.Error())
		}
	}(rebootDelay)

	return rpcmessages.ErrorResponse{Success: true}
}

// GetBaseUpdateProgress returns the Base update progress.
// This RPC should only be called by the app after receiving an OpBaseUpdateProgressChanged notification.
func (middleware *Middleware) GetBaseUpdateProgress() rpcmessages.GetBaseUpdateProgressResponse {
	return middleware.baseUpdateProgress
}

// UpdateBase executes a over-the-air Base update. The version is to be passed as an argument.
// This is archived by running the `bbb-cmd.sh mender-update install <version>` command.
// The current update download progress is read from stdout, parsed and saved as the current state (BaseUpdateState).
// Every time the `BaseUpdateState` of the middleware changes a websocket notification is emitted to the App backend.
// Once the download is complete and the update is applied without errors a Base reboot is scheduled to be executed in 5 seconds.
// The call returns ErrorResponse.Success when a reboot has been scheduled.
func (middleware *Middleware) UpdateBase(args rpcmessages.UpdateBaseArgs) rpcmessages.ErrorResponse {
	log.Println("Starting the Base Update process.")
	// don't allow another update while the states are either downloading, applying or rebooting
	if middleware.baseUpdateProgress.State == rpcmessages.UpdateDownloading ||
		middleware.baseUpdateProgress.State == rpcmessages.UpdateApplying ||
		middleware.baseUpdateProgress.State == rpcmessages.UpdateRebooting {
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "Could not start the update process. A Base update is already in progress or the Base has to be rebooted.",
			Code:    rpcmessages.ErrorMenderUpdateAlreadyInProgress,
		}
	}

	cmd := exec.Command(middleware.environment.GetBBBCmdScript(), "mender-update", "install", args.Version)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Could not get the StdoutPipe to read command progress from: %s", err.Error())
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "Could not start the update process. Please see the Middleware log for more detail.",
			Code:    rpcmessages.ErrorMenderUpdateInstallFailed,
		}
	}
	defer func() {
		err := stdout.Close()
		if err != nil {
			log.Printf("Could not close the stdout pipe %s", err)
		}
	}()
	stdoutScanner := bufio.NewScanner(stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Could not get the StderrPipe to read command progress from: %s", err.Error())
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "Could not start the update process. Please see the Middleware log for more detail.",
			Code:    rpcmessages.ErrorMenderUpdateInstallFailed,
		}
	}
	defer func() {
		err := stderr.Close()
		if err != nil {
			log.Printf("Could not close the stderr pipe %s", err)
		}
	}()
	stderrScanner := bufio.NewScanner(stderr)

	err = cmd.Start()
	if err != nil {
		log.Printf("Could not run the Base update command: %s", err.Error())
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: "Could not start the update process. Please see the Middleware log for more detail.",
			Code:    rpcmessages.ErrorMenderUpdateInstallFailed,
		}
	}

	errOutLines := make([]string, 0)
	// This goroutine uses the bufio.Scanner to .Scan() `stderr` lines.
	// This is done in a goroutine, since .Scan() blocks when there is no input available.
	// Every line read is appended to errOutLines, a string slice with stderr lines.
	// The goroutine exits once EOF for `stderr` is reached.
	// Once `stdout` reaches EOF the `stderr` pipe is closed (see below).
	go func() {
		for {
			hasReadSomething := stderrScanner.Scan()
			if hasReadSomething {
				// Lines written to stderr are captured in errOutLines and processed if os.Wait returns an error
				lineErr := stderrScanner.Text()
				errOutLines = append(errOutLines, strings.TrimSuffix(lineErr, "\n"))
			} else {
				if stderrScanner.Err() != nil {
					log.Printf("GetBaseUpdateProgress: Could not read from stderr scanner: %s", stderrScanner.Err())
					err := stderr.Close()
					if err != nil {
						log.Printf("Could not close the stderr pipe %s", err)
					}
					return
				}
				// When scanner.Scan() returns `false` and scanner.Err() is `nil` then EOF of `stderr` is reached. The goroutine exits.
				return
			}
		}
	}()

	for {
		hasReadSomething := stdoutScanner.Scan()
		if hasReadSomething {
			lineOut := stdoutScanner.Text()
			log.Println(lineOut)
			containsProgressUpdateInfo, percentage, downloadedKiB := parseBaseUpdateStdout(lineOut)
			if containsProgressUpdateInfo {
				middleware.baseUpdateProgress.ProgressPercentage = percentage
				middleware.baseUpdateProgress.ProgressDownloadedKiB = downloadedKiB
				if percentage < 98 {
					middleware.setBaseUpdateStateAndNotify(rpcmessages.UpdateDownloading)
				} else {
					// switch State from `UpdateDownloading` to `UpdateApplying`
					// This is done at 98% or above since the mender-install script does
					// only log 100% after applying the update.
					middleware.setBaseUpdateStateAndNotify(rpcmessages.UpdateApplying)
				}
			}
		} else {
			if stdoutScanner.Err() != nil {
				log.Printf("GetBaseUpdateProgress: Could not read from stdout scanner: %s", stdoutScanner.Err())
				middleware.setBaseUpdateStateAndNotify(rpcmessages.UpdateFailed)
				err := stderr.Close()
				if err != nil {
					log.Printf("Could not close the stderr pipe %s", err)
				}
				return rpcmessages.ErrorResponse{
					Success: false,
					Message: "An error occurred while performing the Base update. Please see the Middleware log for more detail.",
					Code:    rpcmessages.ErrorMenderUpdateInstallFailed,
				}
			}
			// When scanner.Scan() returns `false` and scanner.Err() is `nil` then EOF of `stdout` is reached.
			err := stderr.Close()
			if err != nil {
				log.Printf("Could not close the stderr pipe %s", err)
			}
			break
		}
	}

	err = cmd.Wait()
	if err != nil {
		errorCode := handleBBBScriptErrorCode(errOutLines, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorMenderUpdateImageNotMenderEnabled,
			rpcmessages.ErrorMenderUpdateInstallFailed,
			rpcmessages.ErrorMenderUpdateInvalidVersion,
			rpcmessages.ErrorMenderUpdateNoVersion,
		})

		middleware.setBaseUpdateStateAndNotify(rpcmessages.UpdateFailed)
		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(errOutLines, "\n"),
			Code:    errorCode,
		}
	}

	middleware.setBaseUpdateStateAndNotify(rpcmessages.UpdateRebooting)
	resp := middleware.RebootBase()
	if !resp.Success {
		return resp
	}
	return rpcmessages.ErrorResponse{Success: true}
}

// EnableRootLogin enables/disables the ssh login of the root user
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableRootLogin(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable root login: %t' via the config script.\n", toggleAction.ToggleSetting)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "rootlogin"})
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

// EnableSSHPasswordLogin enables/disables the ssh login with a password (in addition to ssh keys)
// and returns a ErrorResponse indicating if the call was successful.
func (middleware *Middleware) EnableSSHPasswordLogin(toggleAction rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse {
	log.Printf("Executing 'Enable password login: %t' via the config script.\n", toggleAction.ToggleSetting)
	out, err := middleware.runBBBConfigScript([]string{determineEnableValue(toggleAction), "sshpwlogin"})
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

// SetLoginPassword sets the system main ssh/login password
func (middleware *Middleware) SetLoginPassword(args rpcmessages.SetLoginPasswordArgs) rpcmessages.ErrorResponse {
	log.Println("Setting a new login password via the config script")
	password := args.LoginPassword

	// Unicode passwords are allowed, but each Unicode rune is only counted as one when comparing the length
	// len("₿") = 3
	// len([]rune("₿")) = 1
	if len([]rune(password)) >= 8 {
		out, err := middleware.runBBBConfigScript([]string{"set", "loginpw", password})
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
		Code:    rpcmessages.ErrorSetLoginPasswordTooShort,
	}
}

// GetBaseInfo returns information about the Base in a GetBaseInfoResponse
func (middleware *Middleware) GetBaseInfo() rpcmessages.GetBaseInfoResponse {
	hostname, err := middleware.redisClient.GetString(redis.BaseHostname)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	middlewareIP, err := middleware.prometheusClient.GetMetricString(prometheus.BaseSystemInfo, "base_ipaddress")
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	middlewarePort := middleware.environment.GetMiddlewarePort()

	middlewareTorOnion, err := middleware.redisClient.GetString(redis.MiddlewareOnion)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	isTorEnabled, err := middleware.redisClient.GetBool(redis.TorEnabled)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	isBitcoindListening, err := middleware.redisClient.GetBool(redis.BitcoindListen)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	freeDiskspace, err := middleware.prometheusClient.GetInt(prometheus.BaseFreeDiskspace)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	totalDiskspace, err := middleware.prometheusClient.GetInt(prometheus.BaseTotalDiskspace)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	baseVersion, err := middleware.redisClient.GetString(redis.BaseVersion)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	bitcoindVersion, err := middleware.redisClient.GetString(redis.BitcoindVersion)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	lightningdVersion, err := middleware.redisClient.GetString(redis.LightningdVersion)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	electrsVersion, err := middleware.redisClient.GetString(redis.ElectrsVersion)
	if err != nil {
		errResponse := middleware.redisClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetBaseInfoResponse{ErrorResponse: &errResponse}
	}

	return rpcmessages.GetBaseInfoResponse{
		ErrorResponse: &rpcmessages.ErrorResponse{
			Success: true,
		},
		Status:              "-PLACEHOLDER-", // FIXME: This is a placeholder.
		Hostname:            hostname,
		MiddlewareLocalIP:   middlewareIP,
		MiddlewareLocalPort: middlewarePort,
		MiddlewareTorOnion:  middlewareTorOnion,
		MiddlewareTorPort:   "-PLACEHOLDER-", // FIXME: This is a placeholder.
		IsTorEnabled:        isTorEnabled,
		IsBitcoindListening: isBitcoindListening,
		FreeDiskspace:       freeDiskspace,
		TotalDiskspace:      totalDiskspace,
		BaseVersion:         baseVersion,
		BitcoindVersion:     bitcoindVersion,
		LightningdVersion:   lightningdVersion,
		ElectrsVersion:      electrsVersion,
	}
}

// GetServiceInfo returns the most recent information about services running on the Base such as for example bitcoind, electrs or lightningd.
func (middleware *Middleware) GetServiceInfo() rpcmessages.GetServiceInfoResponse {
	log.Println("Returning the lastest Service info", middleware.serviceInfo)
	return middleware.serviceInfo
}
