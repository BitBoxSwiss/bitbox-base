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
	Start() <-chan Event

	/* --- RPCs --- */
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
	IsBaseUpdateAvailable() rpcmessages.IsBaseUpdateAvailableResponse
	GetBaseInfo() rpcmessages.GetBaseInfoResponse
	GetServiceInfo() rpcmessages.GetServiceInfoResponse
	SetLoginPassword(rpcmessages.SetLoginPasswordArgs) rpcmessages.ErrorResponse
	UserAuthenticate(rpcmessages.UserAuthenticateArgs) rpcmessages.UserAuthenticateResponse
	UserChangePassword(rpcmessages.UserChangePasswordArgs) rpcmessages.ErrorResponse
	SetupStatus() rpcmessages.SetupStatusResponse
	/* --- RPCs end --- */

	ValidateToken(token string) error
	GetMiddlewareVersion() string
}

// Handlers provides a web api
type Handlers struct {
	Router *mux.Router
	//upgrader takes an http request and upgrades the connection with its origin to websocket
	upgrader         websocket.Upgrader
	middleware       Middleware
	middlewareEvents <-chan Event
	eventQueue       []Event

	noiseConfig *noisemanager.NoiseConfig
	nClients    int
	clientsMap  map[int]chan<- []byte
	mu          sync.Mutex
}

// Event represents a Event the middleware passes to the handlers to be send to
// a client.
type Event struct {
	Identifier      []byte
	QueueIfNoClient bool
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
		eventQueue:  make([]Event, 0),
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
		if len(handlers.clientsMap) == 0 && event.QueueIfNoClient {
			handlers.eventQueue = append(handlers.eventQueue, event)
		} else {
			for k := range handlers.clientsMap {
				handlers.clientsMap[k] <- event.Identifier
			}
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
		log.Printf("Failed to write response bytes in version handler %s", err)
	}
}

// wsHandler spawns a new ws client, by upgrading the sent request to websocket.
// It listens indefinitely to events from the middleware and relays them to clients accordingly.
func (handlers *Handlers) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := handlers.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %s", err)
		return
	}

	err = handlers.noiseConfig.InitializeNoise(ws)
	if err != nil {
		log.Printf("Noise connection failed to initialize: %s", err)
		return
	}

	server := rpcserver.NewRPCServer(handlers.middleware)

	handlers.mu.Lock()
	handlers.clientsMap[handlers.nClients] = server.RPCConnection.WriteChan()
	handlers.runWebsocket(ws, server.RPCConnection.ReadChan(), server.RPCConnection.WriteChan(), handlers.nClients)
	handlers.nClients++
	handlers.mu.Unlock()

	go server.Serve()

	handlers.mu.Lock()
	for _, event := range handlers.eventQueue {
		for k := range handlers.clientsMap {
			handlers.clientsMap[k] <- event.Identifier
		}
	}
	handlers.mu.Unlock()
}
