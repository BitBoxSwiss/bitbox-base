// Package handlers implements an api for the bitbox-wallet-app to talk to.
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Middleware provides an interface to the middleware package.
type Middleware interface {
	// Start triggers the main middleware event loop that emits events to be caught by the handlers.
	Start() <-chan *middleware.SampleInfo
	// GetSystemEnv returns a system Environment instance containing host system services information.
	GetSystemEnv() system.Environment
}

// Handlers provides a web api
type Handlers struct {
	Router *mux.Router
	//upgrader takes an http request and upgrades the connection with its origin to websocket
	upgrader   websocket.Upgrader
	middleware Middleware
	//TODO(TheCharlatan): In future this event should have a generic interface (thus only containing raw json)
	middlewareEvents <-chan *middleware.SampleInfo
}

// NewHandlers returns a handler instance.
func NewHandlers(middlewareInstance Middleware) *Handlers {
	router := mux.NewRouter()

	handlers := &Handlers{
		middleware: middlewareInstance,
		Router:     router,
		// TODO(TheCharlatan): The upgrader should do an origin check before upgrading. This is important later once we introduce authentication.
		upgrader: websocket.Upgrader{},
	}
	handlers.Router.HandleFunc("/", handlers.rootHandler).Methods("GET")
	handlers.Router.HandleFunc("/ws", handlers.wsHandler)
	handlers.Router.HandleFunc("/getenv", handlers.getEnvHandler).Methods("GET")

	handlers.middlewareEvents = handlers.middleware.Start()
	return handlers
}

// TODO(TheCharlatan): Define a better error-response system. In future, this should be the first step in an authentication procedure.
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

// wsHandler spawns a new ws client, by upgrading the sent request to websocket and then starts a serveSampleInfoToClient stream.
func (handlers *Handlers) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := handlers.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error() + " Failed to upgrade connection")
	}

	err = handlers.serveSampleInfoToClient(ws)
	log.Println(err.Error(), " Websocket client disconnected.")
}

// serveSampleInfoToClient takes a single connected ws client and streams data to it indefinitely until the client disconnected, or a websocket error forces it to return.
func (handlers *Handlers) serveSampleInfoToClient(ws *websocket.Conn) error {
	var i = 0
	for {
		i++
		val := <-handlers.middlewareEvents
		// TODO(TheCharlatan): When middleware events starts streaming more generic data, this should become json
		blockinfo := fmt.Sprintf("%d %f %d %s", val.Blocks, val.Difficulty, i, val.LightningAlias)
		log.Println(blockinfo)
		err := ws.WriteMessage(websocket.TextMessage, []byte(blockinfo))
		if err != nil {
			log.Println(err.Error() + " Unexpected websocket error")
			ws.Close()
			return err
		}
	}
}
