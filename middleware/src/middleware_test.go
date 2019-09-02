package middleware_test

import (
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/stretchr/testify/require"
)

// setupTestMiddleware middleware returns a middleware setup with testing arguments
func setupTestMiddleware() *middleware.Middleware {
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
	- the relative location is differen depending here the tests are run from
	*/
	argumentMap["bbbConfigScript"] = "/bin/echo"
	argumentMap["bbbCmdScript"] = "/bin/echo"

	testMiddleware := middleware.NewMiddleware(argumentMap)

	return testMiddleware
}

func TestSystemEnvResponse(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	systemEnvResponse := testMiddleware.SystemEnv()

	require.Equal(t, systemEnvResponse.ElectrsRPCPort, "18442")
	require.Equal(t, systemEnvResponse.Network, "testnet")
}

func TestSampleInfo(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	sampleInfo := testMiddleware.SampleInfo()
	emptySampleInfo := rpcmessages.SampleInfoResponse{
		Blocks:         0,
		Difficulty:     0.0,
		LightningAlias: "disconnected",
	}

	require.Equal(t, sampleInfo, emptySampleInfo)
}

func TestVerificationProgress(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	verificationProgress := testMiddleware.VerificationProgress()
	emptyVerificationProgress := rpcmessages.VerificationProgressResponse{
		Blocks:               0,
		Headers:              0,
		VerificationProgress: 0.0,
	}

	require.Equal(t, verificationProgress, emptyVerificationProgress)
}

func TestResyncBitcoin(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	response := testMiddleware.ResyncBitcoin()
	require.Equal(t, response.Success, true)
	require.Equal(t, response.Message, "")
	require.Equal(t, response.Code, "")
}

func TestReindexBitcoin(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	response := testMiddleware.ReindexBitcoin()
	require.Equal(t, response.Success, true)
	require.Equal(t, response.Message, "")
	require.Equal(t, response.Code, "")
}

func TestMountFlashdrive(t *testing.T) {
	testMiddleware := setupTestMiddleware()
	response := testMiddleware.MountFlashdrive()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, "", response.Code)
}

func TestUnmountFlashdrive(t *testing.T) {
	testMiddleware := setupTestMiddleware()
	response := testMiddleware.UnmountFlashdrive()
	require.Equal(t, true, response.Success)
	require.Equal(t, "", response.Message)
	require.Equal(t, "", response.Code)
}

func TestBackupHSMSecret(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	response := testMiddleware.BackupHSMSecret()
	require.Equal(t, response.Success, true)
	require.Equal(t, response.Message, "")
	require.Equal(t, response.Code, "")
}

func TestBackupSysconfig(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	response := testMiddleware.BackupSysconfig()
	require.Equal(t, response.Success, true)
	require.Equal(t, response.Message, "")
	require.Equal(t, response.Code, "")
}

func TestRestore(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test sysconfig arg for Restore() ---*/
	restoreSysconfig, errSysconfig := testMiddleware.Restore(rpcmessages.RestoreSysConfig)

	require.Equal(t, restoreSysconfig.Success, true)
	require.Equal(t, restoreSysconfig.Message, "restore sysconfig \n")
	require.NoError(t, errSysconfig)

	/* --- test hsm secret arg for Restore() ---*/
	restoreHSMSecret, errHSMSecret := testMiddleware.Restore(rpcmessages.RestoreHSMSecret)

	require.Equal(t, restoreHSMSecret.Success, true)
	require.Equal(t, restoreHSMSecret.Message, "restore hsm_secret \n")
	require.NoError(t, errHSMSecret)

	/* --- test an unknown arg for Restore() ---*/
	restoreUnknown, errUnknown := testMiddleware.Restore(-1)

	require.Equal(t, restoreUnknown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, restoreUnknown.Message, "Method -1 not supported for Restore().")
	require.Error(t, errUnknown)
}

func TestUserAuthenticate(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* test login should fail for every user/password, when dummyIsBaseSetup == false */
	notInitalizedArgs := rpcmessages.UserAuthenticateArgs{Username: "dev", Password: "dev"}
	authenticateNotInitalized := testMiddleware.UserAuthenticate(notInitalizedArgs)

	require.Equal(t, false, authenticateNotInitalized.Success)
	require.Equal(t, "authentication unsuccessful", authenticateNotInitalized.Message)
	require.Equal(t, false, testMiddleware.DummyIsBaseSetup())

	/* test initial admin login with dummyAdminPassword. Should fail because dummyIsBaseSetup == false  */
	adminDummyPWArgs := rpcmessages.UserAuthenticateArgs{Username: "admin", Password: testMiddleware.DummyAdminPassword()}
	authenticateAdminDummyPW := testMiddleware.UserAuthenticate(adminDummyPWArgs)

	require.Equal(t, false, authenticateAdminDummyPW.Success)
	require.Equal(t, "authentication unsuccessful", authenticateAdminDummyPW.Message)
	require.Equal(t, false, testMiddleware.DummyIsBaseSetup())

	/* test initial admin login, should succeed because dummyIsBaseSetup == false  */
	adminArgs := rpcmessages.UserAuthenticateArgs{Username: "admin", Password: "ICanHasPassword?"}
	authenticateAdmin := testMiddleware.UserAuthenticate(adminArgs)

	require.Equal(t, true, authenticateAdmin.Success)
	require.Equal(t, false, testMiddleware.DummyIsBaseSetup())

	// change admin password to "abc123def", which sets dummyIsBaseSetup = true
	response := testMiddleware.UserChangePassword(rpcmessages.UserChangePasswordArgs{Username: "admin", NewPassword: "abc123def"})
	require.Equal(t, true, response.Success)

	/* test login dev/dev should succeed now, because dummyIsBaseSetup == true */
	devArgs := rpcmessages.UserAuthenticateArgs{Username: "dev", Password: "dev"}
	authenticateDev := testMiddleware.UserAuthenticate(devArgs)

	require.Equal(t, true, authenticateDev.Success)
	require.Equal(t, true, testMiddleware.DummyIsBaseSetup())

	/* test initial admin login, should fail now because dummyIsBaseSetup == true  */
	authenticateAdmin2 := testMiddleware.UserAuthenticate(adminArgs)

	require.Equal(t, false, authenticateAdmin2.Success, false)
	require.Equal(t, "authentication unsuccessful", authenticateAdmin2.Message)
	require.Equal(t, true, testMiddleware.DummyIsBaseSetup(), true)

	/* test initial admin login with dummyAdminPassword. Should succeed now, because dummyIsBaseSetup == true  */
	adminDummyPW2Args := rpcmessages.UserAuthenticateArgs{Username: "admin", Password: testMiddleware.DummyAdminPassword()}
	authenticateAdminDummyPW2 := testMiddleware.UserAuthenticate(adminDummyPW2Args)

	require.Equal(t, true, authenticateAdminDummyPW2.Success)
	require.Equal(t, true, testMiddleware.DummyIsBaseSetup())

	/* test invalid login (with a invalid username) */
	invalidNameArgs := rpcmessages.UserAuthenticateArgs{Username: "InvalidUserName", Password: ""}
	authenticateInvalidName := testMiddleware.UserAuthenticate(invalidNameArgs)

	require.Equal(t, false, authenticateInvalidName.Success)
	require.Equal(t, "authentication unsuccessful", authenticateInvalidName.Message)

	/* test invalid login (empty username and password) */
	emptyArgs := rpcmessages.UserAuthenticateArgs{Username: "", Password: ""}
	authenticateEmpty := testMiddleware.UserAuthenticate(emptyArgs)

	require.Equal(t, false, authenticateEmpty.Success)
	require.Equal(t, "authentication unsuccessful", authenticateEmpty.Message)
}

func TestUserChangePassword(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* test valid password change */
	validArgs := rpcmessages.UserChangePasswordArgs{Username: "notAdmin", NewPassword: "12345678"}
	changepasswordValid := testMiddleware.UserChangePassword(validArgs)

	require.Equal(t, true, changepasswordValid.Success)
	require.Equal(t, false, testMiddleware.DummyIsBaseSetup())

	/* test admin password change, this should set dummyIsBaseSetup == true */
	newPassword := "123qwert567"
	adminChangeArgs := rpcmessages.UserChangePasswordArgs{Username: "admin", NewPassword: newPassword}
	changepasswordAdminChange := testMiddleware.UserChangePassword(adminChangeArgs)

	require.Equal(t, true, changepasswordAdminChange.Success)
	require.Equal(t, true, testMiddleware.DummyIsBaseSetup())

	/* test invalid password change (to short, needs to be 7 chars) */
	invalidArgs := rpcmessages.UserChangePasswordArgs{NewPassword: "1234567"}
	changepasswordInvalid := testMiddleware.UserChangePassword(invalidArgs)

	require.Equal(t, false, changepasswordInvalid.Success)
	require.Equal(t, "password change unsuccessful (too short)", changepasswordInvalid.Message)

	/* test empty password change */
	emptyArgs := rpcmessages.UserChangePasswordArgs{}
	changepasswordEmpty := testMiddleware.UserChangePassword(emptyArgs)

	require.Equal(t, false, changepasswordEmpty.Success)
	require.Equal(t, "password change unsuccessful (too short)", changepasswordEmpty.Message)
}
func TestGetHostname(t *testing.T) {
	testMiddleware := setupTestMiddleware()
	response := testMiddleware.GetHostname()

	require.Equal(t, true, response.Success)
	require.Equal(t, "get hostname ", response.Hostname)
}

func TestSetHostname(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* test normal hostname */
	validArgs1 := rpcmessages.SetHostnameArgs{Hostname: "bitbox-base-satoshi"}
	response1 := testMiddleware.SetHostname(validArgs1)
	require.Equal(t, true, response1.Success)
	require.Empty(t, response1.Message)

	/* test special char hostname */
	validArgs2 := rpcmessages.SetHostnameArgs{Hostname: "a.hostname-with.allowed.-....special.chars"}
	response2 := testMiddleware.SetHostname(validArgs2)
	require.Equal(t, true, response2.Success)
	require.Empty(t, response2.Message)

	/* test a long and valid 64 char hostname */
	validArgs3 := rpcmessages.SetHostnameArgs{Hostname: "a.loooooooooooooooooooooooooooooooooooooooooong.64-char.hostname"}
	response3 := testMiddleware.SetHostname(validArgs3)
	require.Equal(t, true, response3.Success)
	require.Empty(t, response3.Message)

	/* test a long and invalid 65 char hostname */
	invalidArgs1 := rpcmessages.SetHostnameArgs{Hostname: "a.tooooo.loooooooooooooooooooooooooooooooooooong.65-char.hostname"}
	response4 := testMiddleware.SetHostname(invalidArgs1)
	require.Equal(t, false, response4.Success)
	require.Equal(t, "invalid hostname", response4.Message)

	/* test an invalid UPPERCASE letter hostname */
	invalidArgs2 := rpcmessages.SetHostnameArgs{Hostname: "Bitbox"}
	response5 := testMiddleware.SetHostname(invalidArgs2)
	require.Equal(t, false, response5.Success)
	require.Equal(t, "invalid hostname", response5.Message)
}
