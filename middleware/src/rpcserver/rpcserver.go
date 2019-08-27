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
	Backup(rpcmessages.BackupArgs) (rpcmessages.GenericResponse, error)
	Restore(rpcmessages.RestoreArgs) (rpcmessages.GenericResponse, error)
	GetHostname() rpcmessages.GetHostnameResponse
	SetHostname(rpcmessages.SetHostnameArgs) rpcmessages.ErrorResponse
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

// Backup sends the middleware's GenericResponse over rpc
// Args given can specify e.g. a sysconfig backup or a hsm_secret backup
func (server *RPCServer) Backup(args *rpcmessages.BackupArgs, reply *rpcmessages.GenericResponse) error {
	var err error
	*reply, err = server.middleware.Backup(*args)
	log.Printf("sent reply %v: ", reply)
	return err
}

// Restore sends the middleware's GenericResponse over rpc
// Args given can specify e.g. a sysconfig restore or a hsm_secret restore
func (server *RPCServer) Restore(args *rpcmessages.RestoreArgs, reply *rpcmessages.GenericResponse) error {
	var err error
	*reply, err = server.middleware.Restore(*args)
	log.Printf("sent reply %v: ", reply)
	return err
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

// SetHostname sends the middleware's ErrorResponse over rpc
// The argument given specifys the hostname to be set
func (server *RPCServer) SetHostname(args *rpcmessages.SetHostnameArgs, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.SetHostname(*args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetHostname sends the middleware's GetHostnameResponse over rpc
// The GetHostnameResponse includes the current system hostname
func (server *RPCServer) GetHostname(dummyArg bool, reply *rpcmessages.GetHostnameResponse) error {
	*reply = server.middleware.GetHostname()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// Serve starts a gob rpc server
func (server *RPCServer) Serve() {
	rpc.ServeConn(server.RPCConnection)
}
