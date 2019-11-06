package rpcserver_test

import (
	"net/rpc"
	"sync"
	"testing"

	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	rpcserver "github.com/digitalbitbox/bitbox-base/middleware/src/rpcserver"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcserver/mocks"

	"github.com/stretchr/testify/require"
)

func getToggleSettingArgs() rpcmessages.ToggleSettingArgs {
	return rpcmessages.ToggleSettingArgs{ToggleSetting: true}
}

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
	middlewareMock  *mocks.Middleware
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

	// The mocks are generated with the following command in rpcserver.go:
	//go:generate mockery -name Middleware
	testingRPCServer.middlewareMock = &mocks.Middleware{}

	testingRPCServer.rpcServer = rpcserver.NewRPCServer(testingRPCServer.middlewareMock)
	testingRPCServer.serverWriteChan = testingRPCServer.rpcServer.RPCConnection.WriteChan()
	testingRPCServer.serverReadChan = testingRPCServer.rpcServer.RPCConnection.ReadChan()

	go testingRPCServer.rpcServer.Serve()

	testingRPCServer.client = rpc.NewClient(&rpcConn{readChan: testingRPCServer.clientReadChan, writeChan: testingRPCServer.clientWriteChan})

	// To test the rpcserver, the mocked middleware functions need to accept and return some values.
	testingRPCServer.middlewareMock.On("ValidateToken", "").Return(nil)
	testingRPCServer.middlewareMock.On("SystemEnv").Return(rpcmessages.GetEnvResponse{})
	testingRPCServer.middlewareMock.On("ResyncBitcoin").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("ReindexBitcoin").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("BackupSysconfig").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("BackupHSMSecret").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("SetHostname", rpcmessages.SetHostnameArgs{Hostname: "bitbox-base-test"}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("RestoreSysconfig").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("RestoreHSMSecret").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableTor", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableTorMiddleware", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableTorElectrs", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableTorSSH", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableClearnetIBD", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("ShutdownBase").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("RebootBase").Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableRootLogin", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("EnableSSHPasswordLogin", rpcmessages.ToggleSettingArgs{ToggleSetting: true}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("GetBaseInfo").Return(rpcmessages.GetBaseInfoResponse{})
	testingRPCServer.middlewareMock.On("SetLoginPassword", rpcmessages.SetLoginPasswordArgs{}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("UserAuthenticate", rpcmessages.UserAuthenticateArgs{}).Return(
		rpcmessages.UserAuthenticateResponse{ErrorResponse: &rpcmessages.ErrorResponse{Success: true}},
	)
	testingRPCServer.middlewareMock.On("UserChangePassword", rpcmessages.UserChangePasswordArgs{}).Return(rpcmessages.ErrorResponse{Success: true})
	testingRPCServer.middlewareMock.On("IsBaseUpdateAvailable").Return(rpcmessages.IsBaseUpdateAvailableResponse{ErrorResponse: &rpcmessages.ErrorResponse{Success: true}})

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
	// We pass the generic auth request to rpc's that do not take an argument besides the authentication token.
	authArg := rpcmessages.AuthGenericRequest{}

	var systemEnvReply rpcmessages.GetEnvResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetSystemEnv", authArg, &systemEnvReply)

	var reindexBitcoinReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ReindexBitcoin", authArg, &reindexBitcoinReply)
	require.Equal(t, true, reindexBitcoinReply.Success)

	var resyncBitcoinReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ResyncBitcoin", authArg, &resyncBitcoinReply)
	require.Equal(t, true, resyncBitcoinReply.Success)

	setHostnameArg := rpcmessages.SetHostnameArgs{Hostname: "bitbox-base-test"}
	setHostnameReply := rpcmessages.ErrorResponse{}
	testingRPCServer.RunRPCCall(t, "RPCServer.SetHostname", setHostnameArg, &setHostnameReply)
	require.Equal(t, true, setHostnameReply.Success)

	var backupSysconfigReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.BackupSysconfig", authArg, &backupSysconfigReply)
	require.Equal(t, true, backupSysconfigReply.Success)

	var backupHSMSecretReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.BackupHSMSecret", authArg, &backupHSMSecretReply)
	require.Equal(t, true, backupSysconfigReply.Success)

	var restoreSysconfigReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.RestoreSysconfig", authArg, &restoreSysconfigReply)
	require.Equal(t, true, restoreSysconfigReply.Success)

	var restoreHSMSecretReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.RestoreHSMSecret", authArg, &restoreHSMSecretReply)
	require.Equal(t, true, restoreSysconfigReply.Success)

	var enableTorReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTor", getToggleSettingArgs(), &enableTorReply)
	require.Equal(t, true, enableTorReply.Success)

	var enableTorMiddlewareReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorMiddleware", getToggleSettingArgs(), &enableTorMiddlewareReply)
	require.Equal(t, true, enableTorMiddlewareReply.Success)

	var enableTorElectrsReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorElectrs", getToggleSettingArgs(), &enableTorElectrsReply)
	require.Equal(t, true, enableTorElectrsReply.Success)

	var enableTorSSHReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableTorSSH", getToggleSettingArgs(), &enableTorSSHReply)
	require.Equal(t, true, enableTorSSHReply.Success)

	var enableClearnetIBDReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableClearnetIBD", getToggleSettingArgs(), &enableClearnetIBDReply)
	require.Equal(t, true, enableClearnetIBDReply.Success)

	var enableRootLoginReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableRootLogin", getToggleSettingArgs(), &enableRootLoginReply)
	require.Equal(t, true, enableRootLoginReply.Success)

	var enableSSHPasswordLoginReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.EnableSSHPasswordLogin", getToggleSettingArgs(), &enableSSHPasswordLoginReply)
	require.Equal(t, true, enableSSHPasswordLoginReply.Success)

	userAuthenticateArg := rpcmessages.UserAuthenticateArgs{Username: "", Password: ""}
	var userAuthenticateReply rpcmessages.UserAuthenticateResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.UserAuthenticate", userAuthenticateArg, &userAuthenticateReply)
	require.Equal(t, true, userAuthenticateReply.ErrorResponse.Success)

	userChangePasswordArg := rpcmessages.UserChangePasswordArgs{Username: "", NewPassword: ""}
	var userChangePasswordReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.UserChangePassword", userChangePasswordArg, &userChangePasswordReply)
	require.Equal(t, true, userChangePasswordReply.Success)

	var shutdownBaseReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ShutdownBase", authArg, &shutdownBaseReply)
	require.Equal(t, true, shutdownBaseReply.Success)

	var rebootBaseReply rpcmessages.ErrorResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.RebootBase", authArg, &rebootBaseReply)
	require.Equal(t, true, rebootBaseReply.Success)

	var IsBaseUpdateAvailableReply rpcmessages.IsBaseUpdateAvailableResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.IsBaseUpdateAvailable", authArg, &IsBaseUpdateAvailableReply)
	require.Equal(t, true, IsBaseUpdateAvailableReply.ErrorResponse.Success)

	/*
		This can't be unit tested until there is a Prometheus mock.
			var baseInfoReply rpcmessages.GetBaseInfoResponse
			testingRPCServer.RunRPCCall(t, "RPCServer.GetBaseInfo", authArg, &baseInfoReply)
			require.Equal(t, true, baseInfoReply.ErrorResponse.Success)
	*/
}
