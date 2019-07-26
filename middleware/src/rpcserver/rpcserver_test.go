package rpcserver_test

import (
	"net/rpc"
	"sync"
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
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

func TestRPCServer(t *testing.T) {
	argumentMap := make(map[string]string)
	argumentMap["bitcoinRPCUser"] = "user"
	argumentMap["bitcoinRPCPassword"] = "password"
	argumentMap["bitcoinRPCPort"] = "8332"
	argumentMap["lightningRPCPath"] = "/home/bitcoin/.lightning"
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"
	argumentMap["bbbConfigScript"] = "/home/bitcoin/script.sh"
	middlewareInstance := middleware.NewMiddleware(argumentMap)

	rpcServer := rpcserver.NewRPCServer(middlewareInstance)
	serverWriteChan := rpcServer.RPCConnection.WriteChan()
	serverReadChan := rpcServer.RPCConnection.ReadChan()

	go rpcServer.Serve()

	clientWriteChan := make(chan []byte)
	clientReadChan := make(chan []byte)
	client := rpc.NewClient(&rpcConn{readChan: clientReadChan, writeChan: clientWriteChan})

	request := 1
	var reply middleware.GetEnvResponse
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := client.Call("RPCServer.GetSystemEnv", request, &reply)
		require.NoError(t, err)
	}()
	msgRequest := <-clientWriteChan
	serverReadChan <- msgRequest
	msgResponse := <-serverWriteChan
	t.Logf("response message %s", string(msgResponse))
	// Cut off the significant Byte in the response
	clientReadChan <- msgResponse[1:]
	wg.Wait()
	t.Logf("reply: %v", reply)
	require.Equal(t, "testnet", reply.Network)
	require.Equal(t, "18442", reply.ElectrsRPCPort)

	var resyncReply middleware.ResyncBitcoinResponse
	wg = sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := client.Call("RPCServer.ResyncBitcoin", request, &resyncReply)
		require.NoError(t, err)
	}()

	msgRequest = <-clientWriteChan
	serverReadChan <- msgRequest
	msgResponse = <-serverWriteChan
	t.Logf("Resync Bitcoin Response %q", string(msgResponse))
	// Cut off the significant Byte in the response
	clientReadChan <- msgResponse[1:]
	wg.Wait()
	require.Equal(t, false, resyncReply.Success)
}
