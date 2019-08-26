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

func (testRPC *TestingRPCServer) RunRPCCall(t *testing.T, method string, request int, reply interface{}) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	switch reply {
	case rpcmessages.GetEnvResponse{}:
		go func() {
			defer wg.Done()
			err := testRPC.client.Call(method, request, &rpcmessages.GetEnvResponse{})
			require.NoError(t, err)
		}()
	case rpcmessages.ResyncBitcoinResponse{}:
		go func() {
			defer wg.Done()
			err := testRPC.client.Call(method, request, &rpcmessages.ResyncBitcoinResponse{})
			require.NoError(t, err)
		}()
	case rpcmessages.SampleInfoResponse{}:
		go func() {
			defer wg.Done()
			err := testRPC.client.Call(method, request, &rpcmessages.SampleInfoResponse{})
			require.NoError(t, err)
		}()
	case rpcmessages.VerificationProgressResponse{}:
		go func() {
			defer wg.Done()
			err := testRPC.client.Call(method, request, &rpcmessages.VerificationProgressResponse{})
			require.NoError(t, err)
		}()
	default:
	}
	msgRequest := <-testRPC.clientWriteChan
	testRPC.serverReadChan <- msgRequest
	msgResponse := <-testRPC.serverWriteChan
	// Cut off the significant Byte in the response
	testRPC.clientReadChan <- msgResponse[1:]
	wg.Wait()
	t.Logf("reply: %v", reply)
}

func TestRPCServer(t *testing.T) {
	testingRPCServer := NewTestingRPCServer()
	request := 1
	var systemEnvReply rpcmessages.GetEnvResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetSystemEnv", request, systemEnvReply)

	var resyncReply rpcmessages.ResyncBitcoinResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.ResyncBitcoin", request, resyncReply)

	var sampleInfoReply rpcmessages.SampleInfoResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetSampleInfo", request, sampleInfoReply)

	var verificationProgressReply rpcmessages.VerificationProgressResponse
	testingRPCServer.RunRPCCall(t, "RPCServer.GetVerificationProgress", request, verificationProgressReply)
}
