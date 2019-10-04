// Package handlers implements an api for the bitbox-wallet-app to talk to. It also takes care of running the noise encryption.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	noisemanager "github.com/digitalbitbox/bitbox-base/middleware/src/noise"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcserver"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Middleware provides an interface to the middleware package.
type Middleware interface {
	// Start triggers the main middleware event loop that emits events to be caught by the handlers.
	Start() <-chan []byte

	/* --- RPCs --- */
	SystemEnv() rpcmessages.GetEnvResponse
	SampleInfo() rpcmessages.SampleInfoResponse
	ResyncBitcoin() rpcmessages.ErrorResponse
	ReindexBitcoin() rpcmessages.ErrorResponse
	BackupSysconfig() rpcmessages.ErrorResponse
	BackupHSMSecret() rpcmessages.ErrorResponse
	GetHostname() rpcmessages.GetHostnameResponse
	SetHostname(rpcmessages.SetHostnameArgs) rpcmessages.ErrorResponse
	RestoreSysconfig() rpcmessages.ErrorResponse
	RestoreHSMSecret() rpcmessages.ErrorResponse
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
	/* --- RPCs end --- */

	GetMiddlewareVersion() string
}

// Handlers provides a web api
type Handlers struct {
	Router *mux.Router
	//upgrader takes an http request and upgrades the connection with its origin to websocket
	upgrader         websocket.Upgrader
	middleware       Middleware
	middlewareEvents <-chan []byte

	noiseConfig *noisemanager.NoiseConfig
	nClients    int
	clientsMap  map[int]chan<- []byte
	mu          sync.Mutex
}

// NewHandlers returns a handler instance.
func NewHandlers(middlewareInstance Middleware, dataDir string) *Handlers {
	router := mux.NewRouter()

	handlers := &Handlers{
		middleware:  middlewareInstance,
		Router:      router,
		upgrader:    websocket.Upgrader{},
		noiseConfig: noisemanager.NewNoiseConfig(dataDir),
		nClients:    0,
		clientsMap:  make(map[int]chan<- []byte),
	}

	handlers.Router.HandleFunc("/", handlers.rootHandler).Methods("GET")
	handlers.Router.HandleFunc("/version", handlers.versionHandler).Methods("GET")
	handlers.Router.HandleFunc("/ws", handlers.wsHandler)
	handlers.middlewareEvents = handlers.middleware.Start()

	go handlers.listenEvents()
	return handlers
}

func (handlers *Handlers) listenEvents() {
	for {
		event := <-handlers.middlewareEvents
		handlers.mu.Lock()
		for k := range handlers.clientsMap {
			handlers.clientsMap[k] <- event
		}
		handlers.mu.Unlock()
	}
}

func (handlers *Handlers) removeClient(clientID int) {
	handlers.mu.Lock()
	delete(handlers.clientsMap, clientID)
	handlers.mu.Unlock()
}

// rootHandler provides an endpoint to indicate that the middleware is online and able to handle requests.
func (handlers *Handlers) rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK!!\n"))
	if err != nil {
		log.Println(err.Error() + " Failed to write response bytes in root handler")
	}
}

func (handlers *Handlers) versionHandler(w http.ResponseWriter, r *http.Request) {
	type version struct {
		Version string `json:"version"`
	}

	versionString := handlers.middleware.GetMiddlewareVersion()
	v := version{Version: versionString}
	jsonResponse, err := json.Marshal(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Println(err.Error() + " Failed to write response bytes in version handler")
	}
}

// wsHandler spawns a new ws client, by upgrading the sent request to websocket.
// It listens indefinitely to events from the middleware and relays them to clients accordingly.
func (handlers *Handlers) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := handlers.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error() + " Failed to upgrade connection")
	}

	err = handlers.noiseConfig.InitializeNoise(ws)
	if err != nil {
		log.Println(err.Error() + "Noise connection failed to initialize")
		return
	}
	server := rpcserver.NewRPCServer(handlers.middleware)

	handlers.mu.Lock()
	handlers.clientsMap[handlers.nClients] = server.RPCConnection.WriteChan()
	handlers.runWebsocket(ws, server.RPCConnection.ReadChan(), server.RPCConnection.WriteChan(), handlers.nClients)
	handlers.nClients++
	handlers.mu.Unlock()
	go server.Serve()
}
