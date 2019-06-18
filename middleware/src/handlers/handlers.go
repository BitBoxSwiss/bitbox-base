// Package handlers implements an api for the bitbox-wallet-app to talk to. It also takes care of running the noise encryption.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	noisemanager "github.com/digitalbitbox/bitbox-base/middleware/src/noise"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Middleware provides an interface to the middleware package.
type Middleware interface {
	// Start triggers the main middleware event loop that emits events to be caught by the handlers.
	Start() <-chan interface{}
	// GetSystemEnv returns a system Environment instance containing host system services information.
	GetSystemEnv() system.Environment
}

// Handlers provides a web api
type Handlers struct {
	Router *mux.Router
	//upgrader takes an http request and upgrades the connection with its origin to websocket
	upgrader   websocket.Upgrader
	middleware Middleware
	//TODO(TheCharlatan): Starting from the generic interface, flesh out restrictive types over time as the code implements more services.
	middlewareEvents <-chan interface{}

	noiseConfig *noisemanager.NoiseConfig
}

// NewHandlers returns a handler instance.
func NewHandlers(middlewareInstance Middleware, dataDir string) *Handlers {
	router := mux.NewRouter()

	handlers := &Handlers{
		middleware:  middlewareInstance,
		Router:      router,
		upgrader:    websocket.Upgrader{},
		noiseConfig: noisemanager.NewNoiseConfig(dataDir),
	}
	handlers.Router.HandleFunc("/", handlers.rootHandler).Methods("GET")
	handlers.Router.HandleFunc("/ws", handlers.wsHandler)
	handlers.Router.HandleFunc("/getenv", handlers.getEnvHandler).Methods("GET")

	handlers.middlewareEvents = handlers.middleware.Start()
	return handlers
}

// rootHandler provides an endpoint to indicate that the middleware is online and able to handle requests.
func (handlers *Handlers) rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK!!\n"))
	if err != nil {
		log.Println(err.Error() + " Failed to write response bytes in root handler")
	}
}

func (handlers *Handlers) getEnvHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(handlers.middleware.GetSystemEnv())
	if err != nil {
		log.Println(err.Error() + " Failed to serialize GetSystemEnv data to json in getEnvHandler")
		http.Error(w, "Something went wrong, I cannot read these hieroglyphs.", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println(err.Error() + " Failed to write response bytes in getNetwork handler")
		http.Error(w, "Something went wrong, I cannot read these hieroglyphs", http.StatusInternalServerError)
		return
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

	sendChan, _, _, remoteHasQuitChan := handlers.runWebsocket(ws)
	go func() {
		for {
			select {
			case <-remoteHasQuitChan:
				return
			case event := <-handlers.middlewareEvents:
				log.Println("sending middleware event")
				data, err := json.Marshal(event)
				if err != nil {
					log.Println(err.Error() + " Failed to serialize data to json for runWebsocket")
					return
				}
				sendChan <- data
			}
		}
	}()

}
