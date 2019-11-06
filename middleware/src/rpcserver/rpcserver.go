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

// generate mocks for the interface that can be used for testing:
//go:generate mockery -name Middleware

// Middleware provides an interface to the middleware package.
type Middleware interface {
	SystemEnv() rpcmessages.GetEnvResponse
	ResyncBitcoin() rpcmessages.ErrorResponse
	ReindexBitcoin() rpcmessages.ErrorResponse
	BackupSysconfig() rpcmessages.ErrorResponse
	BackupHSMSecret() rpcmessages.ErrorResponse
	SetHostname(rpcmessages.SetHostnameArgs) rpcmessages.ErrorResponse
	RestoreSysconfig() rpcmessages.ErrorResponse
	RestoreHSMSecret() rpcmessages.ErrorResponse
	EnableTor(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	EnableTorMiddleware(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	EnableTorElectrs(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	EnableTorSSH(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	EnableClearnetIBD(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	ShutdownBase() rpcmessages.ErrorResponse
	RebootBase() rpcmessages.ErrorResponse
	EnableRootLogin(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	EnableSSHPasswordLogin(rpcmessages.ToggleSettingArgs) rpcmessages.ErrorResponse
	UpdateBase(rpcmessages.UpdateBaseArgs) rpcmessages.ErrorResponse
	GetBaseUpdateProgress() rpcmessages.GetBaseUpdateProgressResponse
	IsBaseUpdateAvaliable() rpcmessages.IsBaseUpdateAvailableResponse
	GetBaseInfo() rpcmessages.GetBaseInfoResponse
	GetServiceInfo() rpcmessages.GetServiceInfoResponse
	SetLoginPassword(rpcmessages.SetLoginPasswordArgs) rpcmessages.ErrorResponse
	UserAuthenticate(rpcmessages.UserAuthenticateArgs) rpcmessages.UserAuthenticateResponse
	UserChangePassword(rpcmessages.UserChangePasswordArgs) rpcmessages.ErrorResponse
	SetupStatus() rpcmessages.SetupStatusResponse

	//Authentication:
	ValidateToken(token string) error
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

// Serve starts a gob rpc server
func (server *RPCServer) Serve() {
	rpc.ServeConn(server.RPCConnection)
}

func (server *RPCServer) formulateJWTError(name string) rpcmessages.ErrorResponse {
	log.Printf("received rpc request to %s with invalid json web token", name)
	return rpcmessages.ErrorResponse{
		Success: false,
		Message: "JSON web token validation failed",
		Code:    rpcmessages.JSONWebTokenInvalid,
	}
}

// GetSetupStatus send the middleware's setup status as a SetupStatusResponse over rpc.
func (server *RPCServer) GetSetupStatus(dummyArg bool, reply *rpcmessages.SetupStatusResponse) error {
	*reply = server.middleware.SetupStatus()
	log.Printf("sent reply %v: ", reply)
	return nil
}

/* --- Middleware RPCs start here --- */

// GetSystemEnv sends the middleware's GetEnvResponse over rpc
func (server *RPCServer) GetSystemEnv(args rpcmessages.AuthGenericRequest, reply *rpcmessages.GetEnvResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = rpcmessages.GetEnvResponse{}
		log.Printf("received rpc request to GetSystemEnv with invalid json web token")
		return nil
	}

	*reply = server.middleware.SystemEnv()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ReindexBitcoin sends the middleware's ErrorResponse over rpc
func (server *RPCServer) ReindexBitcoin(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("ReindexBitcoin")
		return nil
	}

	*reply = server.middleware.ReindexBitcoin()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ResyncBitcoin sends the middleware's ErrorResponse over rpc
func (server *RPCServer) ResyncBitcoin(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("ResyncBitcoin")
		return nil
	}

	*reply = server.middleware.ResyncBitcoin()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// BackupSysconfig sends the middleware's ErrorResponse over rpc
func (server *RPCServer) BackupSysconfig(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("BackupSysconfig")
		return nil
	}

	*reply = server.middleware.BackupSysconfig()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// BackupHSMSecret sends the middleware's ErrorResponse over rpc
func (server *RPCServer) BackupHSMSecret(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("BackupHSMSecret")
		return nil
	}

	*reply = server.middleware.BackupHSMSecret()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// RestoreSysconfig sends the middleware's ErrorResponse over rpc
func (server *RPCServer) RestoreSysconfig(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("RestoreSysconfig")
		return nil
	}

	*reply = server.middleware.RestoreSysconfig()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// RestoreHSMSecret sends the middleware's ErrorResponse over rpc
func (server *RPCServer) RestoreHSMSecret(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("RestoreHSMSecret")
		return nil
	}

	*reply = server.middleware.RestoreHSMSecret()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// UserAuthenticate sends the middleware's ErrorResponse over rpc
// Args given specify the username and the password
func (server *RPCServer) UserAuthenticate(args *rpcmessages.UserAuthenticateArgs, reply *rpcmessages.UserAuthenticateResponse) error {
	*reply = server.middleware.UserAuthenticate(*args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// UserChangePassword sends the middleware's ErrorResponse over rpc
// The Arg given specify the username and the new password
func (server *RPCServer) UserChangePassword(args *rpcmessages.UserChangePasswordArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("UserChangePassword")
		return nil
	}

	*reply = server.middleware.UserChangePassword(*args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// SetHostname sends the middleware's ErrorResponse over rpc
// The argument given specifies the hostname to be set
func (server *RPCServer) SetHostname(args *rpcmessages.SetHostnameArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("setHostname")
		return nil
	}

	*reply = server.middleware.SetHostname(*args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTor enables/disables the tor.service and configures bitcoind and lightningd.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTor(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("EnableTor")
		return nil
	}

	*reply = server.middleware.EnableTor(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTorMiddleware enables/disables the tor hidden service for the middleware.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTorMiddleware(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("EnableTorMiddleware")
		return nil
	}

	*reply = server.middleware.EnableTorMiddleware(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTorElectrs enables/disables the tor hidden service for Electrs.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTorElectrs(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("EnableTorElectrs")
		return nil
	}

	*reply = server.middleware.EnableTorElectrs(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableTorSSH enables/disables the tor hidden service for SSH.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableTorSSH(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("EnableTorSSH")
		return nil
	}

	*reply = server.middleware.EnableTorSSH(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableClearnetIBD enables/disables the tor hidden service for SSH.
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableClearnetIBD(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("EnableClearnetIBD")
		return nil
	}

	*reply = server.middleware.EnableClearnetIBD(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// ShutdownBase sends the middleware's ErrorResponse over rpc
// The RPC calls the bbb-cmd.sh script which initialtes a `shutdown now`
func (server *RPCServer) ShutdownBase(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("ShutdownBase")
		return nil
	}

	*reply = server.middleware.ShutdownBase()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// RebootBase sends the middleware's ErrorResponse over rpc
// The RPC calls the bbb-cmd.sh script which initialtes a `reboot`
func (server *RPCServer) RebootBase(args rpcmessages.AuthGenericRequest, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("RebootBase")
		return nil
	}

	*reply = server.middleware.RebootBase()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableRootLogin enables/disables the ssh login of the root user
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableRootLogin(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("EnableRootLogin")
		return nil
	}

	*reply = server.middleware.EnableRootLogin(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// EnableSSHPasswordLogin enables/disables the ssh login with a password (in addition to ssh keys)
// The boolean argument passed is used to for enabling and disabling.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) EnableSSHPasswordLogin(args rpcmessages.ToggleSettingArgs, reply *rpcmessages.ErrorResponse) error {
	*reply = server.middleware.EnableSSHPasswordLogin(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// SetLoginPassword sets the system main ssh/login password
// Passwords have to be at least 8 chars in length.
// For Unicode passwords the number of unicode chars is counted and not the byte count.
// It sends the middleware's ErrorResponse over rpc.
func (server *RPCServer) SetLoginPassword(args rpcmessages.SetLoginPasswordArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		*reply = server.formulateJWTError("SetLoginPassword")
		return nil
	}

	*reply = server.middleware.SetLoginPassword(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetBaseInfo sends the middleware's GetBaseInfoResponse over rpc.
// This includes information about the Base and the Middleware.
func (server *RPCServer) GetBaseInfo(args rpcmessages.AuthGenericRequest, reply *rpcmessages.GetBaseInfoResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		errorResponse := server.formulateJWTError("GetBaseInfo")
		*reply = rpcmessages.GetBaseInfoResponse{ErrorResponse: &errorResponse}
		return nil
	}

	*reply = server.middleware.GetBaseInfo()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetServiceInfo sends the middleware's GetServiceInfoResponse over rpc.
// This includes information about the Base and the Middleware.
func (server *RPCServer) GetServiceInfo(args rpcmessages.AuthGenericRequest, reply *rpcmessages.GetServiceInfoResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		errorResponse := server.formulateJWTError("GetServiceInfo")
		*reply = rpcmessages.GetServiceInfoResponse{ErrorResponse: &errorResponse}
		return nil
	}

	*reply = server.middleware.GetServiceInfo()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// UpdateBase updates the Base image and sends a ErrorResponse over RPC
func (server *RPCServer) UpdateBase(args rpcmessages.UpdateBaseArgs, reply *rpcmessages.ErrorResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		errorResponse := server.formulateJWTError("UpdateBase")
		*reply = errorResponse
		return nil
	}

	*reply = server.middleware.UpdateBase(args)
	log.Printf("sent reply %v: ", reply)
	return nil
}

// GetBaseUpdateProgress sends a GetBaseUpdateProgressResponse over RPC
func (server *RPCServer) GetBaseUpdateProgress(args rpcmessages.AuthGenericRequest, reply *rpcmessages.GetBaseUpdateProgressResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		errorResponse := server.formulateJWTError("GetBaseupdateProgress")
		*reply = rpcmessages.GetBaseUpdateProgressResponse{ErrorResponse: &errorResponse}
		return nil
	}

	*reply = server.middleware.GetBaseUpdateProgress()
	log.Printf("sent reply %v: ", reply)
	return nil
}

// IsBaseUpdateAvaliable sends a IsBaseUpdateAvailableResponse over RPC
func (server *RPCServer) IsBaseUpdateAvaliable(args rpcmessages.AuthGenericRequest, reply *rpcmessages.IsBaseUpdateAvailableResponse) error {
	err := server.middleware.ValidateToken(args.Token)
	if err != nil {
		errorResponse := server.formulateJWTError("IsBaseUpdateAvaliable")
		*reply = rpcmessages.IsBaseUpdateAvailableResponse{ErrorResponse: &errorResponse}
		return nil
	}

	*reply = server.middleware.IsBaseUpdateAvaliable()
	log.Printf("IsBaseUpdateAvaliable reply: %v\n", reply)
	return nil
}

/* --- Middleware RPCs end here --- */
