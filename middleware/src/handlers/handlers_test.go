package handlers_test

import (
	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/configuration"
	"github.com/digitalbitbox/bitbox-base/middleware/src/handlers"
	"github.com/stretchr/testify/require"

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

// setupTestMiddleware middleware returns a middleware setup with testing arguments
func setupTestMiddleware(t *testing.T) *middleware.Middleware {
	/* The config and cmd script are mocked with /bin/echo which just returns
	the passed arguments. The real scripts can't be used here, because
	- the absolute location of those is different on each host this is run on
	- the relative location is different depending here the tests are run from
	*/
	const echoBinaryPath string = "/bin/echo"
	const (
		bbbCmdScript              string = echoBinaryPath
		bbbConfigScript           string = echoBinaryPath
		bbbSystemctlScript        string = echoBinaryPath
		electrsRPCPort            string = "18442"
		imageUpdateInfoURL        string = "https://shiftcrypto.ch/updates/base.json"
		middlewarePort            string = "8085"
		middlewareVersion         string = "0.0.1"
		network                   string = "testnet"
		notificationNamedPipePath string = "/tmp/middleware-notification.pipe"
		prometheusURL             string = "http://localhost:9090"
		redisMock                 bool   = true // Important: mock redis in the unit tests
		redisPort                 string = "6379"
	)

	config := configuration.NewConfiguration(
		configuration.Args{
			BBBCmdScript:              bbbCmdScript,
			BBBConfigScript:           bbbConfigScript,
			BBBSystemctlScript:        bbbSystemctlScript,
			ElectrsRPCPort:            electrsRPCPort,
			ImageUpdateInfoURL:        imageUpdateInfoURL,
			MiddlewarePort:            middlewarePort,
			MiddlewareVersion:         middlewareVersion,
			Network:                   network,
			NotificationNamedPipePath: notificationNamedPipePath,
			PrometheusURL:             prometheusURL,
			RedisMock:                 redisMock,
			RedisPort:                 redisPort,
		},
	)

	testMiddleware, err := middleware.NewMiddleware(config, nil)
	require.NoError(t, err)
	return testMiddleware
}

func TestRootHandler(t *testing.T) {
	middlewareInstance := setupTestMiddleware(t)
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
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"
	argumentMap["bbbConfigScript"] = "/home/bitcoin/script.sh"

	middlewareInstance := setupTestMiddleware(t)
	handlers := handlers.NewHandlers(middlewareInstance, ".base")
	rr := httptest.NewServer(handlers.Router)
	defer rr.Close()

	u := "ws://" + rr.Listener.Addr().String() + "/ws"

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)

	//initialize noise
	_, sendCipher := initializeNoise(ws, t)

	// do not do any pairing verification
	err = ws.WriteMessage(1, []byte("m"))
	require.NoError(t, err)
	//test sending to a non-verified api
	encryptedMessage := sendCipher.Encrypt(nil, nil, []byte(opICanHasPairinVerificashun))
	err = ws.WriteMessage(1, encryptedMessage)
	require.NoError(t, err)

	_, _, err = ws.ReadMessage()
	if err == nil {
		t.Errorf("No unexpected close when close was expected, since writing to an unpaired base")
	}
	err = ws.Close()
	require.NoError(t, err)

	ws, _, err = websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer func() {
		err = ws.Close()
		require.NoError(t, err)
	}()

	//initialize noise
	_, _ = initializeNoise(ws, t)

	//do the pairing verificaion
	err = ws.WriteMessage(1, []byte(opICanHasPairinVerificashun))
	require.NoError(t, err)
	_, responseBytes, err := ws.ReadMessage()
	t.Logf("But this is too much!")
	require.NoError(t, err)
	require.Equal(t, string(responseBytes), string(responseSuccess))
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

	//Ask the BitBoxBase to begin the noise 'XX' handshake
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
