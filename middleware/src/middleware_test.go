package middleware_test

import (
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/stretchr/testify/require"
)

func getToggleSettingArgs(enabled bool) rpcmessages.ToggleSettingArgs {
	return rpcmessages.ToggleSettingArgs{ToggleSetting: enabled}
}

// setupTestMiddleware middleware returns a middleware setup with testing arguments
func setupTestMiddleware(t *testing.T) *middleware.Middleware {
	argumentMap := make(map[string]string)
	argumentMap["bitcoinRPCUser"] = "user"
	argumentMap["bitcoinRPCPassword"] = "password"
	argumentMap["bitcoinRPCPort"] = "8332"
	argumentMap["lightningRPCPath"] = "/home/bitcoin/.lightning"
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"

	/* The config and cmd script are mocked with /bin/echo which just returns
	the passed arguments. The real scripts can't be used here, because
	- the absolute location of those is different on each host this is run on
	- the relative location is different depending here the tests are run from
	*/
	const echoBinaryPath string = "/bin/echo"
	argumentMap["bbbConfigScript"] = echoBinaryPath
	argumentMap["bbbCmdScript"] = echoBinaryPath
	argumentMap["bbbSystemctlScript"] = echoBinaryPath

	testMiddleware, err := middleware.NewMiddleware(argumentMap, true)
	require.NoError(t, err)
	return testMiddleware
}

func TestSystemEnvResponse(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	systemEnvResponse := testMiddleware.SystemEnv()

	require.Equal(t, systemEnvResponse.ElectrsRPCPort, "18442")
	require.Equal(t, systemEnvResponse.Network, "testnet")
}

func TestResyncBitcoin(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.ResyncBitcoin()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestReindexBitcoin(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.ReindexBitcoin()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestBackupHSMSecret(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.BackupHSMSecret()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestBackupSysconfig(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.BackupSysconfig()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message, "")
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestRestoreHSMSecret(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.RestoreHSMSecret()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestRestoreSysconfig(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.RestoreSysconfig()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestEnableTor(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableTor(getToggleSettingArgs(true))
	require.Equal(t, true, responseEnable.Success)
	require.Equal(t, "", responseEnable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableTor(getToggleSettingArgs(false))
	require.Equal(t, true, responseDisable.Success)
	require.Equal(t, "", responseDisable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestEnableTorMiddleware(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableTorMiddleware(getToggleSettingArgs(true))
	require.Equal(t, responseEnable.Success, true)
	require.Equal(t, responseEnable.Message, "")
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableTorMiddleware(getToggleSettingArgs(false))
	require.Equal(t, true, responseDisable.Success)
	require.Equal(t, "", responseDisable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestEnableTorElectrs(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableTorElectrs(getToggleSettingArgs(true))
	require.Equal(t, true, responseEnable.Success)
	require.Equal(t, "", responseEnable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableTorElectrs(getToggleSettingArgs(false))
	require.Equal(t, true, responseDisable.Success)
	require.Equal(t, "", responseDisable.Message, "")
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestEnableClearnetIBD(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableClearnetIBD(getToggleSettingArgs(true))
	require.Equal(t, true, responseEnable.Success)
	require.Equal(t, "", responseEnable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableClearnetIBD(getToggleSettingArgs(false))
	require.Equal(t, true, responseDisable.Success)
	require.Equal(t, "", responseDisable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestEnableTorSSH(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableTorSSH(getToggleSettingArgs(true))
	require.Equal(t, true, responseEnable.Success)
	require.Equal(t, "", responseEnable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableTorSSH(getToggleSettingArgs(false))
	require.Equal(t, responseDisable.Success, true)
	require.Equal(t, "", responseDisable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestEnableSSHPasswordLogin(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableSSHPasswordLogin(getToggleSettingArgs(true))
	require.Equal(t, true, responseEnable.Success)
	require.Equal(t, "", responseEnable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableSSHPasswordLogin(getToggleSettingArgs(false))
	require.Equal(t, true, responseDisable.Success, true)
	require.Equal(t, "", responseDisable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestEnableRootLogin(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	responseEnable := testMiddleware.EnableRootLogin(getToggleSettingArgs(true))
	require.Equal(t, true, responseEnable.Success)
	require.Equal(t, "", responseEnable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseEnable.Code)

	responseDisable := testMiddleware.EnableRootLogin(getToggleSettingArgs(false))
	require.Equal(t, true, responseDisable.Success, true)
	require.Equal(t, "", responseDisable.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseDisable.Code)
}

func TestSetLoginPassword(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	// test valid login password set
	responseValid := testMiddleware.SetLoginPassword(rpcmessages.SetLoginPasswordArgs{LoginPassword: "iusethispasswordeverywhere"})
	require.Equal(t, true, responseValid.Success)
	require.Equal(t, "", responseValid.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseValid.Code)

	// test invalid (too short) login password set
	responseInvalid := testMiddleware.SetLoginPassword(rpcmessages.SetLoginPasswordArgs{LoginPassword: "shrtone"})
	require.Equal(t, false, responseInvalid.Success)
	require.Equal(t, "The password has to be at least 8 chars. An unicode char is counted as one.", responseInvalid.Message)
	require.Equal(t, rpcmessages.ErrorSetLoginPasswordTooShort, responseInvalid.Code)

	// test 7 unicode's as password (too short)
	responseUnicode7 := testMiddleware.SetLoginPassword(rpcmessages.SetLoginPasswordArgs{LoginPassword: "₿₿₿₿₿₿₿"})
	require.Equal(t, false, responseUnicode7.Success)
	require.Equal(t, "The password has to be at least 8 chars. An unicode char is counted as one.", responseUnicode7.Message)
	require.Equal(t, rpcmessages.ErrorSetLoginPasswordTooShort, responseUnicode7.Code)

	// test 8 unicode's as password (valid)
	responseUnicode8 := testMiddleware.SetLoginPassword(rpcmessages.SetLoginPasswordArgs{LoginPassword: "₿😂🔥🌑🚀📈世界"})
	require.Equal(t, true, responseUnicode8.Success)
	require.Equal(t, "", responseUnicode8.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), responseUnicode8.Code)
}

func TestUserAuthenticate(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	/* test login should fail for every user/password except default admin, when isMiddlewarePasswordSet == false */
	notInitalizedArgs := rpcmessages.UserAuthenticateArgs{Username: "dev", Password: "dev"}
	authenticateNotInitialized := testMiddleware.UserAuthenticate(notInitalizedArgs)

	require.Equal(t, false, authenticateNotInitialized.ErrorResponse.Success)
	require.Equal(t, "authentication unsuccessful, username not found", authenticateNotInitialized.ErrorResponse.Message)
	require.Equal(t, "", authenticateNotInitialized.Token)
	setupStatus := testMiddleware.SetupStatus()
	require.Equal(t, false, setupStatus.MiddlewarePasswordSet)

	/* test initial admin login, should succeed because isMiddlewarePasswordSet == false and the admin password is set to the initial constant */
	adminArgs := rpcmessages.UserAuthenticateArgs{Username: "admin", Password: testMiddleware.InitialAdminPassword()}
	authenticateAdmin := testMiddleware.UserAuthenticate(adminArgs)

	require.Equal(t, true, authenticateAdmin.ErrorResponse.Success)
	setupStatus = testMiddleware.SetupStatus()
	require.Equal(t, false, setupStatus.MiddlewarePasswordSet)
	// validate the returned token
	require.NoError(t, testMiddleware.ValidateToken(authenticateAdmin.Token))

	// change admin password to "abc123def", which sets isMiddlewarePasswordSet = true
	response := testMiddleware.UserChangePassword(rpcmessages.UserChangePasswordArgs{Username: "admin", Password: testMiddleware.InitialAdminPassword(), NewPassword: "abc123def"})
	require.Equal(t, true, response.Success)

	/* test login admin/abc123def should succeed now, because isMiddlewarePasswordSet == true */
	adminChangedArgs := rpcmessages.UserAuthenticateArgs{Username: "admin", Password: "abc123def"}
	authenticateAdminChanged := testMiddleware.UserAuthenticate(adminChangedArgs)

	require.Equal(t, true, authenticateAdminChanged.ErrorResponse.Success)
	require.NoError(t, testMiddleware.ValidateToken(authenticateAdmin.Token))
	setupStatus = testMiddleware.SetupStatus()
	require.Equal(t, true, setupStatus.MiddlewarePasswordSet)

	/* test initial admin login, should fail now because isMiddlewarePasswordSet == true  */
	authenticateAdmin2 := testMiddleware.UserAuthenticate(adminArgs)

	require.Equal(t, false, authenticateAdmin2.ErrorResponse.Success, false)
	require.Equal(t, "authentication unsuccessful, incorrect password", authenticateAdmin2.ErrorResponse.Message)
	require.Equal(t, "", authenticateAdmin2.Token)
	setupStatus = testMiddleware.SetupStatus()
	require.Equal(t, true, setupStatus.MiddlewarePasswordSet, true)

	/* test invalid login (with a invalid username) */
	invalidNameArgs := rpcmessages.UserAuthenticateArgs{Username: "InvalidUserName", Password: ""}
	authenticateInvalidName := testMiddleware.UserAuthenticate(invalidNameArgs)

	require.Equal(t, false, authenticateInvalidName.ErrorResponse.Success)
	require.Equal(t, "authentication unsuccessful, username not found", authenticateInvalidName.ErrorResponse.Message)

	/* test invalid login (empty username and password) */
	emptyArgs := rpcmessages.UserAuthenticateArgs{Username: "", Password: ""}
	authenticateEmpty := testMiddleware.UserAuthenticate(emptyArgs)

	require.Equal(t, false, authenticateEmpty.ErrorResponse.Success)
	require.Equal(t, "authentication unsuccessful, username not found", authenticateEmpty.ErrorResponse.Message)
}

func TestUserChangePassword(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	/* test password change with invalid initial password*/
	validArgs := rpcmessages.UserChangePasswordArgs{Username: "notAdmin", Password: "notTheInitialPassword", NewPassword: "12345678"}
	changepasswordValid := testMiddleware.UserChangePassword(validArgs)

	require.Equal(t, false, changepasswordValid.Success)
	setupStatus := testMiddleware.SetupStatus()
	require.Equal(t, false, setupStatus.MiddlewarePasswordSet)

	/* test admin password change, this should set isMiddlewarePasswordSet == true */
	newPassword := "123qwert567"
	adminChangeArgs := rpcmessages.UserChangePasswordArgs{Username: "admin", Password: testMiddleware.InitialAdminPassword(), NewPassword: newPassword}
	changepasswordAdminChange := testMiddleware.UserChangePassword(adminChangeArgs)

	require.Equal(t, true, changepasswordAdminChange.Success)
	setupStatus = testMiddleware.SetupStatus()
	require.Equal(t, true, setupStatus.MiddlewarePasswordSet)

	/* test invalid password change (too short, needs to be 7 chars) */
	invalidArgs := rpcmessages.UserChangePasswordArgs{NewPassword: "1234567"}
	changepasswordInvalid := testMiddleware.UserChangePassword(invalidArgs)

	require.Equal(t, false, changepasswordInvalid.Success)
	require.Equal(t, "password change unsuccessful, the password needs to be at least 8 characters in length", changepasswordInvalid.Message)

	/* test empty password change */
	emptyArgs := rpcmessages.UserChangePasswordArgs{}
	changepasswordEmpty := testMiddleware.UserChangePassword(emptyArgs)

	require.Equal(t, false, changepasswordEmpty.Success)
	require.Equal(t, "password change unsuccessful, the password needs to be at least 8 characters in length", changepasswordEmpty.Message)
}

func TestSetHostname(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	/* test normal hostname */
	validArgs1 := rpcmessages.SetHostnameArgs{Hostname: "bitbox-base-satoshi"}
	response1 := testMiddleware.SetHostname(validArgs1)
	require.Equal(t, true, response1.Success)
	require.Empty(t, response1.Message)

	/* test a long and valid 24 char hostname */
	validArgs3 := rpcmessages.SetHostnameArgs{Hostname: "a-loong-24-char-hostname"}
	response3 := testMiddleware.SetHostname(validArgs3)
	require.Equal(t, true, response3.Success)
	require.Empty(t, response3.Message)

	/* test a long and invalid 25 char hostname */
	invalidArgs1 := rpcmessages.SetHostnameArgs{Hostname: "too-long-25-char-hostname"}
	response4 := testMiddleware.SetHostname(invalidArgs1)
	require.Equal(t, false, response4.Success)
	require.Equal(t, "invalid hostname", response4.Message)

	/* test an invalid UPPERCASE letter hostname */
	invalidArgs2 := rpcmessages.SetHostnameArgs{Hostname: "Bitbox"}
	response5 := testMiddleware.SetHostname(invalidArgs2)
	require.Equal(t, false, response5.Success)
	require.Equal(t, "invalid hostname", response5.Message)

	/* test a hostname that ends with a minus sign  */
	invalidArgs3 := rpcmessages.SetHostnameArgs{Hostname: "ending-with-"}
	response6 := testMiddleware.SetHostname(invalidArgs3)
	require.Equal(t, false, response6.Success)
	require.Equal(t, "invalid hostname", response6.Message)

	/* test a hostname that starts with a number  */
	invalidArgs4 := rpcmessages.SetHostnameArgs{Hostname: "0-number-start"}
	response7 := testMiddleware.SetHostname(invalidArgs4)
	require.Equal(t, false, response7.Success)
	require.Equal(t, "invalid hostname", response7.Message)
}

func TestShutdownBase(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.ShutdownBase()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestRebootBase(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.RebootBase()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}

func TestFinalizeBackup(t *testing.T) {
	testMiddleware := setupTestMiddleware(t)

	response := testMiddleware.FinalizeSetupWizard()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, rpcmessages.ErrorCode(""), response.Code)
}
