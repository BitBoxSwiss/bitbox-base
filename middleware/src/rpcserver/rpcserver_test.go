package rpcserver_test

import (
	"net/rpc"
	"sync"
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	rpcserver "github.com/digitalbitbox/bitbox-base/middleware/src/rpcserver"

	"github.com/stretchr/testify/require"
)

type rpcConn struct {
	readChan  <-chan []byte
	writeChan chan<- []byte
}

func (conn *rpcConn) Read(p []byte) (n int, err error) {
	return copy(p, <-conn.readChan), nil
}

func (conn *rpcConn) Write(p []byte) (n int, err error) {
	conn.writeChan <- p
	return len(p), nil
}

func (conn *rpcConn) Close() error {
	return nil
}

type TestingRPCServer struct {
	clientWriteChan chan []byte
	clientReadChan  chan []byte
	serverWriteChan chan []byte
	serverReadChan  chan []byte
	client          *rpc.Client
	rpcServer       *rpcserver.RPCServer
}

func NewTestingRPCServer() TestingRPCServer {
	testingRPCServer := TestingRPCServer{
		clientWriteChan: make(chan []byte),
		clientReadChan:  make(chan []byte),
	}
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

	middlewareInstance := middleware.NewMiddleware(argumentMap)

	testingRPCServer.rpcServer = rpcserver.NewRPCServer(middlewareInstance)
	testingRPCServer.serverWriteChan = testingRPCServer.rpcServer.RPCConnection.WriteChan()
	testingRPCServer.serverReadChan = testingRPCServer.rpcServer.RPCConnection.ReadChan()

	go testingRPCServer.rpcServer.Serve()

	testingRPCServer.client = rpc.NewClient(&rpcConn{readChan: testingRPCServer.clientReadChan, writeChan: testingRPCServer.clientWriteChan})
	return testingRPCServer
}

func (testRPC *TestingRPCServer) RunRPCCall(t *testing.T, method string, arg interface{}, reply interface{}) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := testRPC.client.Call(method, arg, reply)
		require.NoError(t, err)
	}()

	msgRequest := <-testRPC.clientWriteChan
	testRPC.serverReadChan <- msgRequest
	msgResponse := <-testRPC.serverWriteChan
	// Cut off the significant Byte in the response
	testRPC.clientReadChan <- msgResponse[1:]
	wg.Wait()
	t.Logf("%s reply: %v", method, reply)
}

func TestRPCServer(t *testing.T) {
	testingRPCServer := NewTestingRPCServer()

	// The RPCs must get an argument passed.
	// We pass a boolean to RPCs that don't need an argument.
	dummyArg := true

	var systemEnvReply rpcmessages.GetEnvResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetSystemEnv", dummyArg, &systemEnvReply)

	var reindexBitcoinReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ReindexBitcoin", dummyArg, &reindexBitcoinReply)
	require.Equal(t, true, reindexBitcoinReply.Success)

	var resyncBitcoinReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ResyncBitcoin", dummyArg, &resyncBitcoinReply)
	require.Equal(t, true, resyncBitcoinReply.Success)

	var sampleInfoReply rpcmessages.SampleInfoResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetSampleInfo", dummyArg, &sampleInfoReply)

	setHostnameArg := rpcmessages.SetHostnameArgs{Hostname: "bitbox-base-test"}
	setHostnameReply := rpcmessages.ErrorResponse{}
	testingRPCServer.RunRPCCall(t, "RPCServer.SetHostname", setHostnameArg, &setHostnameReply)
	require.Equal(t, true, setHostnameReply.Success)

	var getHostnameReply rpcmessages.GetHostnameResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetHostname", dummyArg, &getHostnameReply)
	require.Equal(t, true, getHostnameReply.ErrorResponse.Success)

	var verificationProgressReply rpcmessages.VerificationProgressResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetVerificationProgress", dummyArg, &verificationProgressReply)

	var mountFlashdriveReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.MountFlashdrive", dummyArg, &mountFlashdriveReply)
	require.Equal(t, true, mountFlashdriveReply.Success)

	var unmountFlashdriveReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.UnmountFlashdrive", dummyArg, &unmountFlashdriveReply)
	require.Equal(t, true, unmountFlashdriveReply.Success)

	var backupSysconfigReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.BackupSysconfig", dummyArg, &backupSysconfigReply)
	require.Equal(t, true, backupSysconfigReply.Success)

	var backupHSMSecretReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.BackupHSMSecret", dummyArg, &backupHSMSecretReply)
	require.Equal(t, true, backupSysconfigReply.Success)

	var restoreSysconfigReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.RestoreSysconfig", dummyArg, &restoreSysconfigReply)
	require.Equal(t, true, restoreSysconfigReply.Success)

	var restoreHSMSecretReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.RestoreHSMSecret", dummyArg, &restoreHSMSecretReply)
	require.Equal(t, true, restoreSysconfigReply.Success)

	var enableTorReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTor", true, &enableTorReply)
	require.Equal(t, true, enableTorReply.Success)

	var disableTorReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTor", false, &disableTorReply)
	require.Equal(t, true, disableTorReply.Success)

	var enableTorMiddlewareReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorMiddleware", true, &enableTorMiddlewareReply)
	require.Equal(t, true, enableTorMiddlewareReply.Success)

	var disableTorMiddlewareReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorMiddleware", false, &disableTorMiddlewareReply)
	require.Equal(t, true, disableTorMiddlewareReply.Success)

	var enableTorElectrsReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorElectrs", true, &enableTorElectrsReply)
	require.Equal(t, true, enableTorElectrsReply.Success)

	var disableTorElectrsReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorElectrs", false, &disableTorElectrsReply)
	require.Equal(t, true, disableTorElectrsReply.Success)

	var enableTorSSHReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorSSH", true, &enableTorSSHReply)
	require.Equal(t, true, enableTorSSHReply.Success)

	var disableTorSSHReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorSSH", false, &disableTorSSHReply)
	require.Equal(t, true, disableTorSSHReply.Success)

	var enableClearnetIBDReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableClearnetIBD", true, &enableClearnetIBDReply)
	require.Equal(t, true, enableClearnetIBDReply.Success)

	var disableClearnetIBDReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableClearnetIBD", false, &disableClearnetIBDReply)
	require.Equal(t, true, disableClearnetIBDReply.Success)

	userAuthenticateArg := rpcmessages.UserAuthenticateArgs{Username: "admin", Password: "ICanHasPassword?"}
	var userAuthenticateReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.UserAuthenticate", userAuthenticateArg, &userAuthenticateReply)
	require.Equal(t, true, userAuthenticateReply.Success)

	userChangePasswordArg := rpcmessages.UserChangePasswordArgs{Username: "admin", NewPassword: "longerpassword"}
	var userChangePasswordReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.UserChangePassword", userChangePasswordArg, &userChangePasswordReply)
	require.Equal(t, true, userChangePasswordReply.Success)

	var shutdownBaseReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ShutdownBase", dummyArg, &shutdownBaseReply)
	require.Equal(t, true, shutdownBaseReply.Success)

	var rebootBaseReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.RebootBase", dummyArg, &rebootBaseReply)
	require.Equal(t, true, rebootBaseReply.Success)

}
