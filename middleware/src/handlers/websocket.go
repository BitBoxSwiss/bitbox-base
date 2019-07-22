package handlers

import (
	"log"

	"github.com/gorilla/websocket"
)

const (
	opICanHasPairinVerificashun = byte('v')
)

// runWebsocket sets up loops for sending/receiving, abstracting away the low level details about
// timeouts, clients closing, etc.
// It returns four channels: one to send messages to the client, one which notifies when the
// It takes four arguments, a websocket connection, a read and a write channel.
//
// The goroutines close client upon exit or dues to a send/receive error.
func (handlers *Handlers) runWebsocket(client *websocket.Conn, readChan chan<- []byte, writeChan <-chan []byte, clientID int) {

	const maxMessageSize = 512
	// this channel is used to break the write loop, when the read loop breaks
	closeChan := make(chan struct{})

	readLoop := func() {
		defer func() {
			_ = client.Close()
			handlers.removeClient(clientID)
			close(closeChan)
			log.Printf("Closed Read Loop for client %v", clientID)
		}()
		client.SetReadLimit(maxMessageSize)
		for {
			_, msg, err := client.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Println("Error, websocket closed unexpectedly in the reading loop")
				}
				log.Printf("Error when reading websocket message, exiting read loop, %v", err)
				return
			}
			// check if it is the message to request the pairing
			if len(msg) == 0 {
				log.Println("Error, received a messaged with zero length, dropping it")
				continue
			}
			if msg[0] == opICanHasPairinVerificashun {
				msg = handlers.noiseConfig.CheckVerification()
				err = client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("Error, websocket failed to write channel hash verification message")
				}
				continue
			}

			messageDecrypted, err := handlers.noiseConfig.Decrypt(msg)
			if err != nil {
				log.Println("Error, websocket could not decrypt incoming packages")
				return
			}
			log.Println(string(messageDecrypted))
			readChan <- messageDecrypted
		}
	}

	writeLoop := func() {
		defer func() {
			_ = client.Close()
			handlers.removeClient(clientID)
			log.Printf("Closed Write Loop for %v", clientID)
		}()
		for {
			select {
			case message, ok := <-writeChan:
				if !ok {
					log.Printf("Error receiving from writeChan %q", string(message))
					_ = client.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				err := client.WriteMessage(websocket.TextMessage, handlers.noiseConfig.Encrypt(message))
				if err != nil {
					log.Println("Error, websocket closed unexpectedly in the writing loop")
					_ = client.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
			case <-closeChan:
				log.Println("Read Loop break, closing write loop")
				return
			}
		}
	}

	go readLoop()
	go writeLoop()
}
