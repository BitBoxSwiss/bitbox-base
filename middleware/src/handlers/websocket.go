package handlers

import (
	"log"

	"github.com/gorilla/websocket"
)

// runWebsocket sets up loops for sending/receiving, abstracting away the low level details about
// timeouts, clients closing, etc.
// It returns four channels: one to send messages to the client, one which notifies when the
// client was closed, one to receive messages from the client and one where the base wants
// to close the connection
//
// Closing the weHaveQuit channel makes runWebsocket's goroutines quit.
// The goroutines close client upon exit, due to a send/receive error or when weHaveQuit is closed.
// runWebsocket never closes weHaveQuit. If it receives a websocket closing message, or has an
// error when receiving a message, it will close the remoteHasQuit channel.
func (handlers *Handlers) runWebsocket(client *websocket.Conn) (send chan<- []byte, weHaveQuit chan<- struct{}, receive <-chan []byte, remoteHasQuit <-chan struct{}) {
	const maxMessageSize = 512

	weHaveQuitChan := make(chan struct{})
	remoteHasQuitChan := make(chan struct{})
	sendChan := make(chan []byte)
	receiveChan := make(chan []byte)

	readLoop := func() {
		defer func() {
			close(remoteHasQuitChan)
			_ = client.Close()
		}()
		client.SetReadLimit(maxMessageSize)
		for {
			_, msg, err := client.ReadMessage()
			// check if it is the message to request the pairing
			if string(msg) == "v" {
				msg = handlers.noiseConfig.CheckVerification()
				err = client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("Error, websocket failed to write channel hash verification message")
				}
				continue
			}

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Println("Error, websocket closed unexpectedly in the reading loop")
				}
				break
			}
			messageDecrypted, err := handlers.noiseConfig.Decrypt(msg)
			if err != nil {
				log.Println("Error, websocket could not decrypt incoming packages")
				break
			}
			receiveChan <- messageDecrypted
		}
	}

	writeLoop := func() {
		defer func() {
			_ = client.Close()
		}()
		for {
			select {
			case message, ok := <-sendChan:
				if !ok {
					_ = client.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				err := client.WriteMessage(websocket.TextMessage, handlers.noiseConfig.Encrypt(message))
				if err != nil {
					log.Println("Error, websocket closed unexpectedly in the writing loop")
				}
			case <-weHaveQuitChan:
				_ = client.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("closing websocket connection")
				return
			}
		}
	}

	go readLoop()
	go writeLoop()

	return sendChan, weHaveQuitChan, receiveChan, remoteHasQuitChan
}
