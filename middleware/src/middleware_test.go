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

	resyncBitcoinResponse, err := testMiddleware.ResyncBitcoin(rpcmessages.Resync)

	require.Equal(t, resyncBitcoinResponse.Success, true)
	require.NoError(t, err)
}

func TestFlashdrive(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test check arg for Flashdrive() ---*/
	checkArgs := rpcmessages.FlashdriveArgs{
		Method: rpcmessages.Check,
		Path:   "", // not needed, only needed for mount
	}

	flashdriveCheck, errCheck := testMiddleware.Flashdrive(checkArgs)

	require.Equal(t, flashdriveCheck.Success, true)
	require.Equal(t, flashdriveCheck.Message, "flashdrive check\n")
	require.NoError(t, errCheck)

	/* --- test mount arg for Flashdrive() ---*/
	mountArgs := rpcmessages.FlashdriveArgs{
		Method: rpcmessages.Mount,
		Path:   "/dev/sda",
	}

	flashdriveMount, errMount := testMiddleware.Flashdrive(mountArgs)

	require.Equal(t, flashdriveMount.Success, true)
	require.Equal(t, flashdriveMount.Message, "flashdrive mount /dev/sda\n")
	require.NoError(t, errMount)

	/* --- test unmount arg for Flashdrive() ---*/
	unmountArgs := rpcmessages.FlashdriveArgs{
		Method: rpcmessages.Unmount,
		Path:   "", // not needed, only needed for mount
	}

	flashdriveUnmount, errUnmount := testMiddleware.Flashdrive(unmountArgs)

	require.Equal(t, flashdriveUnmount.Success, true)
	require.Equal(t, flashdriveUnmount.Message, "flashdrive unmount\n")
	require.NoError(t, errUnmount)

	/* --- test an unknown arg for Flashdrive() ---*/
	unknownArgs := rpcmessages.FlashdriveArgs{
		Method: -1,
		Path:   "",
	}

	flashdriveUnknown, errUnknown := testMiddleware.Flashdrive(unknownArgs)

	require.Equal(t, flashdriveUnknown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, flashdriveUnknown.Message, "Method -1 not supported for Flashdrive().")
	require.Error(t, errUnknown)
}

func TestBackup(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test sysconfig arg for Backup() ---*/
	backupSysconfig, errSysconfig := testMiddleware.Backup(rpcmessages.BackupSysConfig)

	require.Equal(t, backupSysconfig.Success, true)
	require.Equal(t, backupSysconfig.Message, "backup sysconfig\n")
	require.NoError(t, errSysconfig)

	/* --- test hsm secret arg for Backup() ---*/
	backupHSMSecret, errHSMSecret := testMiddleware.Backup(rpcmessages.BackupHSMSecret)

	require.Equal(t, backupHSMSecret.Success, true)
	require.Equal(t, backupHSMSecret.Message, "backup hsm_secret\n")
	require.NoError(t, errHSMSecret)

	/* --- test an unknown arg for Backup() ---*/
	backupUnknown, errUnknown := testMiddleware.Backup(-1)

	require.Equal(t, backupUnknown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, backupUnknown.Message, "Method -1 not supported for Backup().")
	require.Error(t, errUnknown)
}

// TestRestore only covers the 'script not found' case.
// We can't know the absolute path of the script, because that depends
// on the system the tests are executed. e.g. Travis path and local path differ.
func TestRestore(t *testing.T) {
	testMiddleware := setupTestMiddleware()

	/* --- test sysconfig arg for Restore() ---*/
	restoreSysconfig, errSysconfig := testMiddleware.Restore(rpcmessages.RestoreSysConfig)

	require.Equal(t, restoreSysconfig.Success, true)
	require.Equal(t, restoreSysconfig.Message, "restore sysconfig\n")
	require.NoError(t, errSysconfig)

	/* --- test hsm secret arg for Restore() ---*/
	restoreHSMSecret, errHSMSecret := testMiddleware.Restore(rpcmessages.RestoreHSMSecret)

	require.Equal(t, restoreHSMSecret.Success, true)
	require.Equal(t, restoreHSMSecret.Message, "restore hsm_secret\n")
	require.NoError(t, errHSMSecret)

	/* --- test an unknown arg for Restore() ---*/
	restoreUnknown, errUnknown := testMiddleware.Restore(-1)

	require.Equal(t, restoreUnknown.Success, false) // should fail, the method -1 is unknown
	require.Equal(t, restoreUnknown.Message, "Method -1 not supported for Restore().")
	require.Error(t, errUnknown)
}
