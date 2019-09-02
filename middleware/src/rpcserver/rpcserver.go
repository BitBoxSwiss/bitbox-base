package rpcserver

import (
	"log"
	"net/rpc"

	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
)

// rpcConn wraps an io.ReadWriteCloser
type rpcConn struct {
	readChan  chan []byte
	writeChan chan []byte
}

// newRPCConn returns an rpcConn struct that can be used as an interface to an io.ReadWriteCloser
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

// Read implements io.ReadWriteCloser
func (conn *rpcConn) Read(p []byte) (n int, err error) {
	message := <-conn.readChan
	return copy(p, message), nil
}

// Write implements io.ReadWriteCloser
func (conn *rpcConn) Write(p []byte) (n int, err error) {
	conn.writeChan <- append([]byte(rpcmessages.OpRPCCall), p...)
	return len(p), nil
}

// Close implements io.ReadWriteCloser. It is just a dummy function.
func (conn *rpcConn) Close() error {
	return nil
}

// Middleware provides an interface to the middleware package.
type Middleware interface {
	SystemEnv() rpcmessages.GetEnvResponse
	ResyncBitcoin(rpcmessages.ResyncBitcoinArgs) (rpcmessages.ResyncBitcoinResponse, error)
	Flashdrive(rpcmessages.FlashdriveArgs) (rpcmessages.GenericResponse, error)
	BackupSysconfig() rpcmessages.ErrorResponse
	BackupHSMSecret() rpcmessages.ErrorResponse
	RestoreSysconfig() rpcmessages.ErrorResponse
	RestoreHSMSecret() rpcmessages.ErrorResponse
	SampleInfo() rpcmessages.SampleInfoResponse
	VerificationProgress() rpcmessages.VerificationProgressResponse
	UserAuthenticate(rpcmessages.UserAuthenticateArgs) rpcmessages.ErrorResponse
	UserChangePassword(rpcmessages.UserChangePasswordArgs) rpcmessages.ErrorResponse
}

// RPCServer provides rpc calls to the middleware
type RPCServer struct {
	middleware    Middleware
	RPCConnection *rpcConn
}

// NewRPCServer returns a new RPCServer
func NewRPCServer(middleware Middleware) *RPCServer {
	server := &RPCServer{
		middleware: middleware,

		//RPCConnection accepts an io.ReadWriteCloser interface from newRPCConn()
		RPCConnection: newRPCConn(),
	}
	err := rpc.Register(server)
	if err != nil {
		log.Println("Unable to register new rpc server")
	}

	return server
}

// GetSystemEnv sends the middleware's GetEnvResponse over rpc
func (server *RPCServer) GetSystemEnv(args int, reply *rpcmessages.GetEnvResponse) error {
	*reply = server.middleware.SystemEnv()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ResyncBitcoin sends the middleware's ResyncBitcoinResponse over rpc
func (server *RPCServer) ResyncBitcoin(args *rpcmessages.ResyncBitcoinArgs, reply *rpcmessages.ResyncBitcoinResponse) error {
	var err error
	*reply, err = server.middleware.ResyncBitcoin(*args)
	log.Printf("sent reply %v: ", reply)
	return err
}

// GetSampleInfo sends the middleware's SampleInfoResponse over rpc
func (server *RPCServer) GetSampleInfo(args int, reply *rpcmessages.SampleInfoResponse) error {
	*reply = server.middleware.SampleInfo()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetVerificationProgress sends the middleware's VerificationProgressResponse over rpc
func (server *RPCServer) GetVerificationProgress(args int, reply *rpcmessages.VerificationProgressResponse) error {
	*reply = server.middleware.VerificationProgress()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// Flashdrive sends the middleware's GenericResponse over rpc
// Args given can specify e.g. a flashdrive check, mount or unmount
func (server *RPCServer) Flashdrive(args *rpcmessages.FlashdriveArgs, reply *rpcmessages.GenericResponse) error {
	var err error
	*reply, err = server.middleware.Flashdrive(*args)
	log.Printf("sent reply %v: ", reply)
	return err
}

// BackupSysconfig sends the middleware's ErrorResponse over rpc
func (server *RPCServer) BackupSysconfig(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.BackupSysconfig()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// BackupHSMSecret sends the middleware's ErrorResponse over rpc
func (server *RPCServer) BackupHSMSecret(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.BackupHSMSecret()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// RestoreSysconfig sends the middleware's ErrorResponse over rpc
func (server *RPCServer) RestoreSysconfig(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.RestoreSysconfig()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// RestoreHSMSecret sends the middleware's ErrorResponse over rpc
func (server *RPCServer) RestoreHSMSecret(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.RestoreHSMSecret()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// UserAuthenticate sends the middleware's ErrorResponse over rpc
// Args given specify the username and the password
func (server *RPCServer) UserAuthenticate(args *rpcmessages.UserAuthenticateArgs, reply *rpcmessages.ErrorResponse) {
	*reply = server.middleware.UserAuthenticate(*args)
	log.Printf("sent reply %v: ", reply)
}

// UserChangePassword sends the middleware's ErrorResponse over rpc
// The Arg given specify the username and the new password
func (server *RPCServer) UserChangePassword(args *rpcmessages.UserChangePasswordArgs, reply *rpcmessages.ErrorResponse) {
	*reply = server.middleware.UserChangePassword(*args)
	log.Printf("sent reply %v: ", reply)
}

// Serve starts a gob rpc server
func (server *RPCServer) Serve() {
	rpc.ServeConn(server.RPCConnection)
}
