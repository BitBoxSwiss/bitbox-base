package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/digitalbitbox/bitbox-base/middleware/src/handlers"
	"github.com/digitalbitbox/bitbox-base/middleware/src/prometheus"
	"github.com/digitalbitbox/bitbox-base/middleware/src/redis"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
)

// The util.go file includes utility functions for the Middleware.
// These are private and called by the middleware RPCs. Utility
// functions like `mountFlashdrive` or `unmountFlashdrive` are
// reused in multiple RPCs.

// mountFlashdrive tries to mount a flashdrive. It first checks if one and
// only one flashdrive is available. If yes, then this flashdrive is mounted
// under /mnt/backup as defined in the bbb-cmd.sh script.
// On error an ErrorResponse is returned containing the necessary data for
// the frontend (not successful, (error) message and (error) code).
func (middleware *Middleware) mountFlashdrive() rpcmessages.ErrorResponse {
	log.Println("Executing a USB flashdrive check via the cmd script")
	outCheck, err := middleware.runBBBCmdScript([]string{"flashdrive", "check"})

	if err != nil {
		errorCode := handleBBBScriptErrorCode(outCheck, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorFlashdriveCheckMultiple,
			rpcmessages.ErrorFlashdriveCheckNone,
		})

		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(outCheck, "\n"),
			Code:    errorCode,
		}
	}

	if len(outCheck) > 1 { // `bbb-cmd.sh flashdrive check` returns only the flashdrive name, if no errors occur
		log.Printf("Warning: The `bbb-cmd.sh flashdrive check` command returned more than one line. Using '%s' as flashdrive name.\n.", outCheck[len(outCheck)-1])
	} else if len(outCheck) == 0 {
		return rpcmessages.ErrorResponse{ // throw an unexpected error if the script does not return anything
			Success: false,
			Message: "An unexpected Error occurred: The output of `bbb-cmd.sh flashdrive check` did not contain a flashdrive name.",
			Code:    rpcmessages.ErrorUnexpected,
		}
	}

	flashDriveName := outCheck[len(outCheck)-1]

	log.Println("Executing a USB flashdrive mount via the cmd script")
	outMount, err := middleware.runBBBCmdScript([]string{"flashdrive", "mount", flashDriveName})
	if err != nil {
		errorCode := handleBBBScriptErrorCode(outMount, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorFlashdriveMountNotFound,
			rpcmessages.ErrorFlashdriveMountNotSupported,
			rpcmessages.ErrorFlashdriveMountNotUnique,
		})

		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(outMount, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// unmountFlashdrive tries to unmount a flashdrive mounted at `/mnt/backup`
// as defined in the bbb-cmd.sh script. If there is no flashdrive mounted,
// an ErrorResponse with the ErrorCode ErrorFlashdriveUnmountNotMounted is
// returned.
func (middleware *Middleware) unmountFlashdrive() rpcmessages.ErrorResponse {
	log.Println("Executing a USB flashdrive unmount via the cmd script")
	out, err := middleware.runBBBCmdScript([]string{"flashdrive", "unmount"})

	if err != nil {
		errorCode := handleBBBScriptErrorCode(out, err, []rpcmessages.ErrorCode{
			rpcmessages.ErrorFlashdriveUnmountNotMounted,
		})

		return rpcmessages.ErrorResponse{
			Success: false,
			Message: strings.Join(out, "\n"),
			Code:    errorCode,
		}
	}

	return rpcmessages.ErrorResponse{Success: true}
}

// runBBBCmdScript runs the bbb-cmd.sh script.
// The script executes commands like for example mounting a USB drive, doing a backup and copying files.
func (middleware *Middleware) runBBBCmdScript(args []string) (outputLines []string, err error) {
	outputLines, err = runCommand(middleware.environment.GetBBBCmdScript(), args)
	return
}

// runBBBConfigScript runs the bbb-config.sh script.
// The script changes the system configuration in redis by setting or unsetting the appropriate keys.
// If necessary the affected services are restarted.
func (middleware *Middleware) runBBBConfigScript(args []string) (outputLines []string, err error) {
	outputLines, err = runCommand(middleware.environment.GetBBBConfigScript(), args)
	return
}

// runCommand runs the passed command and returns the combined stdout and stderr output.
// If the command could not be run, err is not nil.
func runCommand(command string, args []string) (combinedLines []string, err error) {
	cmd := exec.Command(command, args...)
	log.Printf("executing command: %s %s", command, strings.Join(args, " "))

	rawstdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		// no error handling here, just logging
		log.Printf("Error executing command '%s %s': '%s'", command, strings.Join(args, " "), err.Error())
	}

	combined := strings.TrimSuffix(string(rawstdoutStderr), "\n")
	combinedLines = strings.Split(combined, "\n")
	return combinedLines, err
}

func handleBBBScriptErrorCode(outputLines []string, err error, possibleErrors []rpcmessages.ErrorCode) rpcmessages.ErrorCode {
	// There are two possible types of errors handled here:
	// 1.   The script could not be found. -> return ExecutableNotFound
	// 2.   Script exited with `exit status 1`:
	// 	2.1. Script was not run as superuser. ErrorCode ErrorScriptNotSuperuser is expected as the last outputLine.
	// 	2.2. CMD Script was not run with correct parameters. ErrorCode ErrorCmdScriptInvalidArg is expected as the last outputLine.
	// 	2.3. Config Script was not run with correct parameters. ErrorCode ErrorConfigScriptInvalidArg is expected as the last outputLine.
	// 	2.4. One of the `possibleErrors` is expected as the last outputLine.
	// All other errors are unknown and not handled. ErrorUnexpected is returned as a last resort.

	if os.IsNotExist(err) {
		return rpcmessages.ExecutableNotFound
	} else if err.Error() == "exit status 1" {
		if len(outputLines) == 0 {
			log.Println("Error: no log lines provided before exit with error status 1.")
			return rpcmessages.ErrorUnexpected
		}

		outputErrorCode := outputLines[len(outputLines)-1]

		// handling default errors the bbb-cmd and bbb-config scripts can return
		switch outputErrorCode {
		case string(rpcmessages.ErrorScriptNotSuperuser):
			return rpcmessages.ErrorScriptNotSuperuser // Script was not run as superuser.
		case string(rpcmessages.ErrorCmdScriptInvalidArg):
			return rpcmessages.ErrorCmdScriptInvalidArg // Invalid arguments passed to the bbb-cmd.sh script.
		case string(rpcmessages.ErrorConfigScriptInvalidArg):
			return rpcmessages.ErrorConfigScriptInvalidArg // Invalid arguments passed to the bbb-config.sh script.
		}

		// handling specific possible errors the executed part of the script can throw
		// e.g. flashdrive mounting errors when executing a flashdrive mount
		for _, possible := range possibleErrors {
			if outputErrorCode == string(possible) {
				return possible
			}
		}
	}

	log.Printf("Error: unhandled error '%s' with output '%s'", err.Error(), outputLines)
	return rpcmessages.ErrorUnexpected
}

// determineEnableValue returns a string (either "enable" or "disable") used as parameter for the bbb-config.sh script for a given ToggleSettingArgs
func determineEnableValue(enable rpcmessages.ToggleSettingArgs) string {
	if enable.ToggleSetting {
		return "enable"
	}
	return "disable"
}

func (middleware *Middleware) didServiceInfoChange() (changed bool) {
	upToDateServiceInfo := middleware.getServiceInfo()

	// Since the pointer addresses for the ErrorResponses are not equal the == operator can't be used.
	// reflect.DeepEqual() checks if the values at the pointer addresses are equal.
	if !reflect.DeepEqual(upToDateServiceInfo, middleware.serviceInfo) {
		middleware.serviceInfo = upToDateServiceInfo
		log.Println("new serviceInfo", middleware.serviceInfo)
		return true
	}
	return false
}

// getServiceInfo returns a up-to-date GetServiceInfoResponse with information about `bitcoind`, `lightningd` and `electrs`.
func (middleware *Middleware) getServiceInfo() rpcmessages.GetServiceInfoResponse {
	bitcoindBlocks, err := middleware.prometheusClient.GetInt(prometheus.BitcoinBlockCount)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}

	bitcoindHeaders, err := middleware.prometheusClient.GetInt(prometheus.BitcoinHeaderCount)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}

	bitcoindVerificationProgress, err := middleware.prometheusClient.GetFloat(prometheus.BitcoinVerificationProgress)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}

	bitcoindPeers, err := middleware.prometheusClient.GetInt(prometheus.BitcoinPeers)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}

	bitcoindIBDAsInt, err := middleware.prometheusClient.GetInt(prometheus.BitcoinIBD)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}
	bitcoindIBD := bitcoindIBDAsInt == 1

	lightningdBlocks, err := middleware.prometheusClient.GetInt(prometheus.LightningBlocks)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}

	electrsBlocks, err := middleware.prometheusClient.GetInt(prometheus.ElectrsBlocks)
	if err != nil {
		errResponse := middleware.prometheusClient.ConvertErrorToErrorResponse(err)
		return rpcmessages.GetServiceInfoResponse{ErrorResponse: &errResponse}
	}

	return rpcmessages.GetServiceInfoResponse{
		ErrorResponse: &rpcmessages.ErrorResponse{
			Success: true,
		},
		BitcoindBlocks:               bitcoindBlocks,
		BitcoindHeaders:              bitcoindHeaders,
		BitcoindVerificationProgress: bitcoindVerificationProgress,
		BitcoindPeers:                bitcoindPeers,
		BitcoindIBD:                  bitcoindIBD,
		LightningdBlocks:             lightningdBlocks,
		ElectrsBlocks:                electrsBlocks,
	}
}

// parseBaseUpdateStdout parses the output generated by the mender install script.
// This out put can either be general log output or progress output.
// The update progress is saved in the Middleware and an event is emitted to the middleware that indicates a update progress change.
// Update progress output look similar to:
//   ...............................   0% 1024 KiB
func parseBaseUpdateStdout(outputLine string) (containsUpdateProgressInfo bool, percentage int, downloadedKiB int) {
	const prefix string = "................................"
	const suffix string = "KiB"

	if !strings.HasPrefix(outputLine, prefix) || !strings.HasSuffix(outputLine, suffix) {
		// the read line does not contain information about a the update progress
		return false, 0, 0
	}

	strippedProgress := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(outputLine, suffix), prefix))

	splittedProgress := strings.Split(strippedProgress, " ")
	if len(splittedProgress) != 2 {
		log.Printf("parseBaseUpdateStdout: Unexpected string parts in stripped output '%s'", strippedProgress)
		return false, 0, 0
	}

	a := strings.Replace(splittedProgress[0], "%", "", 1)
	percentage, err := strconv.Atoi(a)
	if err != nil {
		log.Printf("parseBaseUpdateStdout: Could not convert '%s' to an integer: %s", a, err)
		return false, 0, 0
	}

	downloadedKiB, err = strconv.Atoi(splittedProgress[1])
	if err != nil {
		log.Printf("parseBaseUpdateStdout: Could not convert '%s' to an integer: %s", splittedProgress[1], err)
		return false, 0, 0
	}

	log.Println(percentage, downloadedKiB)
	return true, percentage, downloadedKiB
}

func (middleware *Middleware) setBaseUpdateStateAndNotify(state rpcmessages.BaseUpdateState) {
	middleware.baseUpdateProgress.State = state
	middleware.events <- handlers.Event{
		Identifier:      []byte(rpcmessages.OpBaseUpdateProgressChanged),
		QueueIfNoClient: true,
	}
}

// checkMiddlewareSetup checks if the middleware password has been set yet and if the user is done with the base
// setup by getting these values from redis and placing them into the middleware struct. Returns an error if the redis lookup fails.
func (middleware *Middleware) checkMiddlewareSetup() error {
	passwordSet, err := middleware.redisClient.GetBool(redis.MiddlewarePasswordSet)
	if err != nil {
		return err
	}
	baseSetupDone, err := middleware.redisClient.GetBool(redis.BaseSetupDone)
	if err != nil {
		return err
	}
	middleware.isMiddlewarePasswordSet = passwordSet
	middleware.isBaseSetupDone = baseSetupDone
	return nil
}

// get userAuthStructure gets the user authentication data from redis and parses it into a map with the username as a key and the UserAuthStruct
// as a value.
func (middleware *Middleware) getAuthStructure() (map[string]UserAuthStruct, error) {
	usersMap := make(map[string]UserAuthStruct)
	// isMiddlewarePasswordSet checks if the base is run the first time.
	authStructureString, err := middleware.redisClient.GetString(redis.MiddlewareAuth)
	if err != nil {
		log.Println("error getting the auth structure string from the redis client")
		return usersMap, err
	}

	err = json.Unmarshal([]byte(authStructureString), &usersMap)
	if err != nil {
		log.Println("Did not receive json from redis's authentication structure")
		return usersMap, err
	}
	return usersMap, nil
}

// getBaseUpdateInfo GETs a JSON file over HTTP which includes information about the update.
func getBaseUpdateInfo(url string) (updateInfo rpcmessages.UpdateInfo, err error) {
	client := http.Client{
		Timeout: 15 * time.Second, // timeout is chosen arbitrary, but should account for (very) slow tor connections.
	}

	response, err := client.Get(url)
	if err != nil {
		return updateInfo, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	updateInfo = rpcmessages.UpdateInfo{}

	err = json.NewDecoder(response.Body).Decode(&updateInfo)
	if err != nil {
		return updateInfo, fmt.Errorf("could not decode JSON response: %s", err)
	}

	return updateInfo, nil
}
