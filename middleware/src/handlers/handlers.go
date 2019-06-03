// Package handlers implements an api for the bitbox-wallet-app to talk to.
package handlers

import (
	"fmt"
	"log"
	"net/http"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Middleware provides an interface to the middleware package.
type Middleware interface {
	Start() <-chan *middleware.SampleInfo
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
