package handlers_test

import (
	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/handlers"
	basemessages "github.com/digitalbitbox/bitbox-base/middleware/src/messages"
	"github.com/stretchr/testify/require"

	"github.com/golang/protobuf/proto"

	"github.com/flynn/noise"

	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

const (
	opICanHasHandShaek          = "h"
	opICanHasPairinVerificashun = "v"
	responseSuccess             = "\x00"
	responseNeedsPairing        = "\x01"
)

func TestRootHandler(t *testing.T) {
	argumentMap := make(map[string]string)
	argumentMap["bitcoinRPCUser"] = "user"
	argumentMap["bitcoinRPCPassword"] = "password"
	argumentMap["bitcoinRPCPort"] = "8332"
	argumentMap["lightningRPCPath"] = "/home/bitcoin/.lightning"
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"
	argumentMap["bbbConfigScript"] = "/home/bitcoin/script.sh"

	middlewareInstance := middleware.NewMiddleware(argumentMap)
	handlers := handlers.NewHandlers(middlewareInstance, ".base")
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handlers.Router.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusOK)
	require.Equal(t, rr.Body.String(), "OK!!\n")
}

func TestWebsocketHandler(t *testing.T) {
	argumentMap := make(map[string]string)
	argumentMap["bitcoinRPCUser"] = "user"
	argumentMap["bitcoinRPCPassword"] = "password"
	argumentMap["bitcoinRPCPort"] = "8332"
	argumentMap["lightningRPCPath"] = "/home/bitcoin/.lightning"
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"
	argumentMap["bbbConfigScript"] = "/home/bitcoin/script.sh"

	middlewareInstance := middleware.NewMiddleware(argumentMap)
	handlers := handlers.NewHandlers(middlewareInstance, ".base")
	rr := httptest.NewServer(handlers.Router)
	defer rr.Close()

	u := "ws://" + rr.Listener.Addr().String() + "/ws"

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)

	//initialize noise
	_, sendCipher := initializeNoise(ws, t)

	//test sending to an unpaired api
	encryptedMessage := sendCipher.Encrypt(nil, nil, []byte(opICanHasPairinVerificashun))
	err = ws.WriteMessage(1, encryptedMessage)
	require.NoError(t, err)

	_, _, err = ws.ReadMessage()
	if err == nil {
		t.Errorf("No unexpected close when close was expected, since writing to an unpaired base")
	}
	ws.Close()

	ws, _, err = websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	//initialize noise
	receiveCipher, sendCipher := initializeNoise(ws, t)

	//do the pairing verificaion
	err = ws.WriteMessage(1, []byte(opICanHasPairinVerificashun))
	require.NoError(t, err)
	_, responseBytes, err := ws.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, string(responseBytes), string(responseSuccess))

	outgoing := &basemessages.BitBoxBaseIn{
		BitBoxBaseIn: &basemessages.BitBoxBaseIn_BaseSystemEnvIn{
			BaseSystemEnvIn: &basemessages.BaseSystemEnvIn{},
		},
	}
	data, err := proto.Marshal(outgoing)
	require.NoError(t, err)
	err = ws.WriteMessage(1, sendCipher.Encrypt(nil, nil, data))
	require.NoError(t, err)
	_, responseBytes, err = ws.ReadMessage()
	require.NoError(t, err)
	_, err = receiveCipher.Decrypt(nil, nil, responseBytes)
	require.NoError(t, err)

	outgoing = &basemessages.BitBoxBaseIn{
		BitBoxBaseIn: &basemessages.BitBoxBaseIn_BaseResyncIn{
			BaseResyncIn: &basemessages.BaseResyncIn{},
		},
	}
	data, err = proto.Marshal(outgoing)
	require.NoError(t, err)
	err = ws.WriteMessage(1, sendCipher.Encrypt(nil, nil, data))
	require.NoError(t, err)
	_, responseBytes, err = ws.ReadMessage()
	require.NoError(t, err)
	_, err = receiveCipher.Decrypt(nil, nil, responseBytes)
	require.NoError(t, err)

}

// initializeNoise sets up a new noise connection. First a fresh keypair is generated if none is locally found.
// Afterwards a XX handshake is performed. This is a three part handshake required to authenticate both parties.
// The resulting pairing code is then displayed to the user to check if it matches what is displayed on the other party's device.
func initializeNoise(client *websocket.Conn, t *testing.T) (*noise.CipherState, *noise.CipherState) {
	cipherSuite := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256)
	kp, err := cipherSuite.GenerateKeypair(rand.Reader)
	require.NoError(t, err)
	handshake, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:   cipherSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		StaticKeypair: kp, //*keypair,
		Prologue:      []byte("Noise_XX_25519_ChaChaPoly_SHA256"),
		Initiator:     true,
	})
	require.NoError(t, err)

	//Ask the BitBox Base to begin the noise 'XX' handshake
	err = client.WriteMessage(1, []byte(opICanHasHandShaek))
	require.NoError(t, err)
	_, responseBytes, err := client.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, string(responseBytes), string(responseSuccess))
	// Do 3 part noise 'XX' handshake.
	msg, _, _, err := handshake.WriteMessage(nil, nil)
	require.NoError(t, err)
	err = client.WriteMessage(websocket.BinaryMessage, msg)
	require.NoError(t, err)
	_, responseBytes, err = client.ReadMessage()
	require.NoError(t, err)
	_, _, _, err = handshake.ReadMessage(nil, responseBytes)
	require.NoError(t, err)
	msg, receiveCipher, sendCipher, err := handshake.WriteMessage(nil, nil)
	require.NoError(t, err)
	err = client.WriteMessage(websocket.BinaryMessage, msg)
	require.NoError(t, err)

	//read the pairing verification request
	_, responseBytes, err = client.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, string(responseBytes), string(responseNeedsPairing))

	return receiveCipher, sendCipher
}
