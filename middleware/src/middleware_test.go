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
	argumentMap["bbbConfigScript"] = "/home/bitcoin/bbb-cmd.sh"
	argumentMap["bbbCmdScript"] = "/home/bitcoin/cmd-script.sh"

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

// TestResyncBitcoin only covers the script not found case.
// We can't know the absolute path of the script, because that depends
// on the system the tests are executed. e.g. Travis path and local path differ.
func TestResyncBitcoin(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	resyncBitcoinResponse, err := testMiddleware.ResyncBitcoin(rpcmessages.Resync)

	require.Equal(t, resyncBitcoinResponse.Success, false)
	require.Error(t, err)
}

// TestFlashdrive only covers the 'script not found' case.
// We can't know the absolute path of the script, because that depends
// on the system the tests are executed. e.g. Travis path and local path differ.
func TestFlashdrive(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test check arg for Flashdrive() ---*/
	checkArgs := rpcmessages.FlashdriveArgs{
		Method: rpcmessages.Check,
		Path:   "", // not needed, only needed for mount
	}

	flashdriveCheck, errCheck := testMiddleware.Flashdrive(checkArgs)

	require.Equal(t, flashdriveCheck.Success, false)
	require.Empty(t, flashdriveCheck.Message) // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errCheck)

	/* --- test mount arg for Flashdrive() ---*/
	mountArgs := rpcmessages.FlashdriveArgs{
		Method: rpcmessages.Mount,
		Path:   "/dev/sda",
	}

	flashdriveMount, errMount := testMiddleware.Flashdrive(mountArgs)

	require.Equal(t, flashdriveMount.Success, false)
	require.Empty(t, flashdriveMount.Message) // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errMount)

	/* --- test unmount arg for Flashdrive() ---*/
	unmountArgs := rpcmessages.FlashdriveArgs{
		Method: rpcmessages.Unmount,
		Path:   "", // not needed, only needed for mount
	}

	flashdriveUnmount, errUnmount := testMiddleware.Flashdrive(unmountArgs)

	require.Equal(t, flashdriveUnmount.Success, false)
	require.Empty(t, flashdriveUnmount.Message) // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errUnmount)

	/* --- test an unkown arg for Flashdrive() ---*/
	unkownArgs := rpcmessages.FlashdriveArgs{
		Method: -1,
		Path:   "",
	}

	flashdriveUnkown, errUnkown := testMiddleware.Flashdrive(unkownArgs)

	require.Equal(t, flashdriveUnkown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, flashdriveUnkown.Message, "Method -1 not supported for Flashdrive().")
	require.Error(t, errUnkown)
}

// TestBackup only covers the 'script not found' case.
// We can't know the absolute path of the script, because that depends
// on the system the tests are executed. e.g. Travis path and local path differ.
func TestBackup(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test sysconfig arg for Backup() ---*/
	backupSysconfig, errSysconfig := testMiddleware.Backup(rpcmessages.BackupSysConfig)

	require.Equal(t, backupSysconfig.Success, false) // should fail, because the script location is invalid
	require.Empty(t, backupSysconfig.Message)        // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errSysconfig)

	/* --- test hsm secret arg for Backup() ---*/
	backupHSMSecret, errHSMSecret := testMiddleware.Backup(rpcmessages.BackupHSMSecret)

	require.Equal(t, backupHSMSecret.Success, false) // should fail, because the script location is invalid
	require.Empty(t, backupHSMSecret.Message)        // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errHSMSecret)

	/* --- test an unknown arg for Backup() ---*/
	backupUnkown, errUnkown := testMiddleware.Backup(-1)

	require.Equal(t, backupUnkown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, backupUnkown.Message, "Method -1 not supported for Backup().")
	require.Error(t, errUnkown)
}

// TestRestore only covers the 'script not found' case.
// We can't know the absolute path of the script, because that depends
// on the system the tests are executed. e.g. Travis path and local path differ.
func TestRestore(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test sysconfig arg for Restore() ---*/
	restoreSysconfig, errSysconfig := testMiddleware.Restore(rpcmessages.RestoreSysConfig)

	require.Equal(t, restoreSysconfig.Success, false) // should fail, because the script location is invalid
	require.Empty(t, restoreSysconfig.Message)        // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errSysconfig)

	/* --- test hsm secret arg for Restore() ---*/
	restoreHSMSecret, errHSMSecret := testMiddleware.Restore(rpcmessages.RestoreHSMSecret)

	require.Equal(t, restoreHSMSecret.Success, false) // should fail, because the script location is invalid
	require.Empty(t, restoreHSMSecret.Message)        // should be empty, because the script should not be found and can't acctually return something
	require.Error(t, errHSMSecret)

	/* --- test an unknown arg for Restore() ---*/
	restoreUnkown, errUnkown := testMiddleware.Restore(-1)

	require.Equal(t, restoreUnkown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, restoreUnkown.Message, "Method -1 not supported for Restore().")
	require.Error(t, errUnkown)
}
