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
	ResyncBitcoin() rpcmessages.ErrorResponse
	ReindexBitcoin() rpcmessages.ErrorResponse
	BackupSysconfig() rpcmessages.ErrorResponse
	BackupHSMSecret() rpcmessages.ErrorResponse
	GetHostname() rpcmessages.GetHostnameResponse
	SetHostname(rpcmessages.SetHostnameArgs) rpcmessages.ErrorResponse
	RestoreSysconfig() rpcmessages.ErrorResponse
	RestoreHSMSecret() rpcmessages.ErrorResponse
	SampleInfo() rpcmessages.SampleInfoResponse
	EnableTor(rpcmessages.ToggleSetting) rpcmessages.ErrorResponse
	EnableTorMiddleware(rpcmessages.ToggleSetting) rpcmessages.ErrorResponse
	EnableTorElectrs(rpcmessages.ToggleSetting) rpcmessages.ErrorResponse
	EnableTorSSH(rpcmessages.ToggleSetting) rpcmessages.ErrorResponse
	EnableClearnetIBD(rpcmessages.ToggleSetting) rpcmessages.ErrorResponse
	ShutdownBase() rpcmessages.ErrorResponse
	RebootBase() rpcmessages.ErrorResponse
	EnableRootLogin(rpcmessages.ToggleSetting) rpcmessages.ErrorResponse
	GetBaseVersion() rpcmessages.GetBaseVersionResponse
	SetRootPassword(rpcmessages.SetRootPasswordArgs) rpcmessages.ErrorResponse
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
func (server *RPCServer) GetSystemEnv(dummyArg bool, reply *rpcmessages.GetEnvResponse) error {
	*reply = server.middleware.SystemEnv()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ReindexBitcoin sends the middleware's ErrorResponse over rpc
func (server *RPCServer) ReindexBitcoin(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.ReindexBitcoin()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ResyncBitcoin sends the middleware's ErrorResponse over rpc
func (server *RPCServer) ResyncBitcoin(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.ResyncBitcoin()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetSampleInfo sends the middleware's SampleInfoResponse over rpc
func (server *RPCServer) GetSampleInfo(dummyArg bool, reply *rpcmessages.SampleInfoResponse) error {
	*reply = server.middleware.SampleInfo()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetVerificationProgress sends the middleware's VerificationProgressResponse over rpc
func (server *RPCServer) GetVerificationProgress(dummyArg bool, reply *rpcmessages.VerificationProgressResponse) error {
	*reply = server.middleware.VerificationProgress()
	log.Printf("sent reply %v: ", reply)
	return nil
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
func (server *RPCServer) UserAuthenticate(args *rpcmessages.UserAuthenticateArgs, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.UserAuthenticate(*args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// UserChangePassword sends the middleware's ErrorResponse over rpc
// The Arg given specify the username and the new password
func (server *RPCServer) UserChangePassword(args *rpcmessages.UserChangePasswordArgs, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.UserChangePassword(*args)
	log.Printf("sent reply %v: ", reply)
	return nil
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

// EnableTor enables/disables the tor.service and configures bitcoind and lightningd.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTor(toggleAction rpcmessages.ToggleSetting, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableTor(toggleAction)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTorMiddleware enables/disables the tor hidden service for the middleware.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTorMiddleware(toggleAction rpcmessages.ToggleSetting, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableTorMiddleware(toggleAction)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTorElectrs enables/disables the tor hidden service for Electrs.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTorElectrs(toggleAction rpcmessages.ToggleSetting, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableTorElectrs(toggleAction)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTorSSH enables/disables the tor hidden service for SSH.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTorSSH(toggleAction rpcmessages.ToggleSetting, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableTorSSH(toggleAction)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableClearnetIBD enables/disables the tor hidden service for SSH.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableClearnetIBD(toggleAction rpcmessages.ToggleSetting, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableClearnetIBD(toggleAction)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ShutdownBase sends the middleware's ErrorResponse over rpc
// The RPC calls the bbb-cmd.sh script which initialtes a `shutdown now`
func (server *RPCServer) ShutdownBase(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.ShutdownBase()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// RebootBase sends the middleware's ErrorResponse over rpc
// The RPC calls the bbb-cmd.sh script which initialtes a `reboot`
func (server *RPCServer) RebootBase(dummyArg bool, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.RebootBase()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableRootLogin enables/disables login via the root user/password.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableRootLogin(toggleAction rpcmessages.ToggleSetting, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableRootLogin(toggleAction)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// SetRootPassword sets the systems root password.
// Passwords have to be at least 8 chars in length.
// For Unicode passwords the number of unicode chars is counted and not the byte count.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) SetRootPassword(args rpcmessages.SetRootPasswordArgs, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.SetRootPassword(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetBaseVersion returns a GetBaseVersionResponse containing the base version.
func (server *RPCServer) GetBaseVersion(dummyArg bool, reply *rpcmessages.GetBaseVersionResponse) error {
	*reply = server.middleware.GetBaseVersion()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// Serve starts a gob rpc server
func (server *RPCServer) Serve() {
	rpc.ServeConn(server.RPCConnection)
}
