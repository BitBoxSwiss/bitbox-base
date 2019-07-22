package rpcserver

import (
	"log"
	"net/rpc"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
)

type rpcConn struct {
	readChan  chan []byte
	writeChan chan []byte
}

func newRPCConn() *rpcConn {
	RPCConn := &rpcConn{
		readChan:  make(chan []byte),
		writeChan: make(chan []byte),
	}
	return RPCConn
}

func (conn *rpcConn) ReadChan() chan []byte {
	return conn.readChan
}

func (conn *rpcConn) WriteChan() chan []byte {
	return conn.writeChan
}

func (conn *rpcConn) Read(p []byte) (n int, err error) {
	message := <-conn.readChan
	return copy(p, message), nil
}

func (conn *rpcConn) Write(p []byte) (n int, err error) {
	conn.writeChan <- append([]byte("r"), p...)
	return len(p), nil
}

func (conn *rpcConn) Close() error {
	return nil
}

// Middleware provides an interface to the middleware package.
type Middleware interface {
	SystemEnv() middleware.GetEnvResponse
	ResyncBitcoin() middleware.ResyncBitcoinResponse
	SampleInfo() middleware.SampleInfoResponse
}

// RPCServer provides rpc calls to the middleware
type RPCServer struct {
	middleware    Middleware
	RPCConnection *rpcConn
}

// NewRPCServer returns a new RPCServer
func NewRPCServer(middleware Middleware) *RPCServer {
	server := &RPCServer{
		middleware:    middleware,
		RPCConnection: newRPCConn(),
	}
	err := rpc.Register(server)
	if err != nil {
		log.Println("Unable to register new rpc server")
	}

	return server
}

// GetSystemEnv sends the middleware's GetEnvResponse over rpc
func (server *RPCServer) GetSystemEnv(args int, reply *middleware.GetEnvResponse) error {
	*reply = server.middleware.SystemEnv()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ResyncBitcoin sends the middleware's ResyncBitcoinResponse over rpc
func (server *RPCServer) ResyncBitcoin(args int, reply *middleware.ResyncBitcoinResponse) error {
	*reply = server.middleware.ResyncBitcoin()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetSampleInfo send the middleware's SampleInfoResponse over rpc
func (server *RPCServer) GetSampleInfo(args int, reply *middleware.SampleInfoResponse) error {
	*reply = server.middleware.SampleInfo()
	log.Printf("sent reply %v: ", reply)
	return nil
}

func (server *RPCServer) Serve() {
	rpc.ServeConn(server.RPCConnection)
}
