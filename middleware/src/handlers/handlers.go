// Package handlers implements an api for the bitbox-wallet-app to talk to. It also takes care of running the noise encryption.
package handlers

import (
	"log"
	"net/http"
	"sync"

	basemessages "github.com/digitalbitbox/bitbox-base/middleware/src/messages"
	noisemanager "github.com/digitalbitbox/bitbox-base/middleware/src/noise"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Middleware provides an interface to the middleware package.
type Middleware interface {
	// Start triggers the main middleware event loop that emits events to be caught by the handlers.
	Start() <-chan []byte
	SystemEnv() []byte
	ResyncBitcoin() []byte
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

// rootHandler provides an endpoint to indicate that the middleware is online and able to handle requests.
func (handlers *Handlers) rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK!!\n"))
	if err != nil {
		log.Println(err.Error() + " Failed to write response bytes in root handler")
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

	sendChan, _, receiveChan, remoteHasQuitChan := handlers.runWebsocket(ws)
	handlers.mu.Lock()
	handlers.clientsMap[0] = sendChan
	handlers.nClients++
	handlers.mu.Unlock()
	go func() {
		for {
			select {
			case message := <-receiveChan:
				incoming := &basemessages.BitBoxBaseIn{}
				if err := proto.Unmarshal(message, incoming); err != nil {
					log.Println("protobuf unmarshal of incoming packet failed")
				}
				switch incoming.BitBoxBaseIn.(type) {
				case *basemessages.BitBoxBaseIn_BaseSystemEnvIn:
					go func() {
						sendChan <- handlers.middleware.SystemEnv()
					}()
				case *basemessages.BitBoxBaseIn_BaseResyncIn:
					go func() {
						sendChan <- handlers.middleware.ResyncBitcoin()
					}()
				}
			case <-remoteHasQuitChan:
				return
			}
		}
	}()
}
