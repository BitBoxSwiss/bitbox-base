// noisemanager gives some useful functions to interact with the noise encryption and decryption, provides verification of the noise channel hash and writes the noise keys to a file
package noisemanager

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/flynn/noise"
	"github.com/gorilla/websocket"
)

const (
	opICanHasHandShaek   = "h"
	responseSuccess      = "\x00"
	responseNeedsPairing = "\x01"
)

type NoiseConfig struct {
	clientStaticPubkey          []byte
	channelHash                 string
	sendCipher, receiveCipher   *noise.CipherState
	pairingVerificationRequired bool
	initialized                 bool
	dataDir                     string
}

func NewNoiseConfig(dataDir string) *NoiseConfig {
	noise := &NoiseConfig{
		dataDir:     dataDir,
		initialized: false,
	}
	return noise
}

// initializeNoise sets up a new noise connection. First a fresh keypair is generated if none is locally found.
// Afterwards a XX handshake is performed. This is a three part handshake required to authenticate both parties.
// The resulting pairing code is then displayed to the user to check if it matches what is displayed on the other party's device.
func (noiseConfig *NoiseConfig) InitializeNoise(ws *websocket.Conn) error {
	cipherSuite := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256)
	keypair := noiseConfig.getMiddlewareNoiseStaticKeypair()
	if keypair == nil {
		if true {
			kp, err := cipherSuite.GenerateKeypair(rand.Reader)
			if err != nil {
				return errors.New("failed to generate a new noise keypair")
			}
			keypair = &kp

			if err := noiseConfig.setMiddlewareNoiseStaticKeypair(keypair); err != nil {
				log.Println("could not store app noise static keypair")
			}
		}
	}
	handshake, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:   cipherSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		StaticKeypair: *keypair,
		Prologue:      []byte("Noise_XX_25519_ChaChaPoly_SHA256"),
		Initiator:     false,
	})
	if err != nil {
		return errors.New("failed to generate a new noise handshake state for the wallet app communication with the BitBox Base")
	}

	// check the websocket connection
	_, responseBytes, err := ws.ReadMessage()
	if err != nil {
		return errors.New("websocket failed to read noise handshake request")
	}
	if string(responseBytes) != string(opICanHasHandShaek) {
		return errors.New("initial response bytes did not match what we were expecting")
	}
	err = ws.WriteMessage(1, []byte(responseSuccess))
	if err != nil {
		return errors.New("websocket failed to write the noise handshake request response")
	}

	// do 3 part noise 'XX' handshake
	_, responseBytes, err = ws.ReadMessage()
	if err != nil {
		return errors.New("websocket failed to read first noise handshake message")
	}
	_, _, _, err = handshake.ReadMessage(nil, responseBytes)
	if err != nil {
		return errors.New("noise failed to read first noise handshake message")
	}
	msg, _, _, err := handshake.WriteMessage(nil, nil)
	if err != nil {
		return errors.New("noise failed to write second noise handshake message")
	}
	err = ws.WriteMessage(1, msg)
	if err != nil {
		return errors.New("websocket failed to write second noise handshake message")
	}
	_, responseBytes, err = ws.ReadMessage()
	if err != nil {
		return errors.New("websocket failed to read third noise handshake message")
	}
	_, noiseConfig.sendCipher, noiseConfig.receiveCipher, err = handshake.ReadMessage(nil, responseBytes)
	if err != nil {
		return errors.New("noise failed to read the third noise handshake message")
	}

	// Check if the user already authenticated the channel binding hash
	noiseConfig.clientStaticPubkey = handshake.PeerStatic()
	if len(noiseConfig.clientStaticPubkey) != 32 {
		return errors.New("expected 32 byte remote static pubkey")
	}
	noiseConfig.pairingVerificationRequired = !noiseConfig.containsClientStaticPubkey(noiseConfig.clientStaticPubkey)

	// If the user has not authenticated, the connected client needs to ask for verification before being able to interact with the base
	if noiseConfig.pairingVerificationRequired {
		err = ws.WriteMessage(websocket.BinaryMessage, []byte(responseNeedsPairing))
		if err != nil {
			return errors.New("websocket failed to write second noise handshake message")
		}

	} else {
		err = ws.WriteMessage(websocket.BinaryMessage, []byte(responseSuccess))
		if err != nil {
			return errors.New("websocket failed to write second noise handshake message")
		}

	}
	channelHashBase32 := base32.StdEncoding.EncodeToString(handshake.ChannelBinding())
	noiseConfig.channelHash = fmt.Sprintf(
		"%s %s\n%s %s",
		channelHashBase32[:5],
		channelHashBase32[5:10],
		channelHashBase32[10:15],
		channelHashBase32[15:20])
	noiseConfig.initialized = true
	return nil
}

func (noiseConfig *NoiseConfig) CheckVerification() []byte {
	// TODO(TheCharlatan) At this point, the channel Hash should be displayed on the screen, with a blocking call.
	// For now, just add a dummy timer, since we do not have a screen yet, and make every verification a success.
	time.Sleep(2 * time.Second)
	err := noiseConfig.addClientStaticPubkey(noiseConfig.clientStaticPubkey)
	if err != nil {
		log.Println("Pairing Successful, but unable to write baseNoiseStaticPubkey to file")
	}
	noiseConfig.pairingVerificationRequired = false
	return []byte(responseSuccess)
}

func (noiseConfig *NoiseConfig) Encrypt(message []byte) []byte {
	if !noiseConfig.initialized {
		return []byte("Error: noise session not initialized")
	}
	if noiseConfig.pairingVerificationRequired {
		message = []byte("Error: encrypted connection not verified")
	}
	return noiseConfig.sendCipher.Encrypt(nil, nil, message)
}

func (noiseConfig *NoiseConfig) Decrypt(message []byte) ([]byte, error) {
	if !noiseConfig.initialized {
		return []byte(""), errors.New("noise not initialized")
	}
	if noiseConfig.pairingVerificationRequired {
		return []byte(""), errors.New("pairing verification has not been done with this client")
	}
	return noiseConfig.receiveCipher.Decrypt(nil, nil, message)
}

const configFilename = "base.json"

type noiseKeypair struct {
	Private []byte `json:"private"`
	Public  []byte `json:"public"`
}

type configuration struct {
	MiddlewareNoiseStaticKeypair *noiseKeypair `json:"appNoiseStaticKeypair"`
	ClientNoiseStaticPubkeys     [][]byte      `json:"deviceNoiseStaticPubkeys"`
}

func (noiseConfig *NoiseConfig) readConfig() *configuration {
	configFile := NewFile(noiseConfig.dataDir, configFilename)
	if !configFile.Exists() {
		return &configuration{}
	}
	var conf configuration
	if err := configFile.ReadJSON(&conf); err != nil {
		return &configuration{}
	}
	return &conf
}

func (noiseConfig *NoiseConfig) storeConfig(conf *configuration) error {
	configFile := NewFile(noiseConfig.dataDir, configFilename)
	return configFile.WriteJSON(conf)
}

func (noiseConfig *NoiseConfig) containsClientStaticPubkey(pubkey []byte) bool {
	for _, configPubkey := range noiseConfig.readConfig().ClientNoiseStaticPubkeys {
		if bytes.Equal(configPubkey, pubkey) {
			return true
		}
	}
	return false
}

func (noiseConfig *NoiseConfig) addClientStaticPubkey(pubkey []byte) error {
	if noiseConfig.containsClientStaticPubkey(pubkey) {
		// Don't add again if already present.
		return nil
	}

	config := noiseConfig.readConfig()
	config.ClientNoiseStaticPubkeys = append(config.ClientNoiseStaticPubkeys, pubkey)
	return noiseConfig.storeConfig(config)
}

func (noiseConfig *NoiseConfig) getMiddlewareNoiseStaticKeypair() *noise.DHKey {
	key := noiseConfig.readConfig().MiddlewareNoiseStaticKeypair
	if key == nil {
		return nil
	}
	return &noise.DHKey{
		Private: key.Private,
		Public:  key.Public,
	}
}

func (noiseConfig *NoiseConfig) setMiddlewareNoiseStaticKeypair(key *noise.DHKey) error {
	config := noiseConfig.readConfig()
	config.MiddlewareNoiseStaticKeypair = &noiseKeypair{
		Private: key.Private,
		Public:  key.Public,
	}
	return noiseConfig.storeConfig(config)
}

// File models a config file in the application's directory.
// Callers can use MiddlewareDir function to obtain the default app config dir.
type File struct {
	dir  string
	name string
}

// NewFile creates a new config file with the given name in a directory dir.
func NewFile(dir, name string) *File {
	return &File{dir: dir, name: name}
}

// Path returns the absolute path to the config file.
func (file *File) Path() string {
	return filepath.Join(file.dir, file.name)
}

// Exists checks whether the file exists with suitable permissions as a file and not as a directory.
func (file *File) Exists() bool {
	info, err := os.Stat(file.Path())
	return err == nil && !info.IsDir()
}

// Remove removes the file.
func (file *File) Remove() error {
	return os.Remove(file.Path())
}

// read reads the config file and returns its data (or an error if the config file does not exist).
func (file *File) read() ([]byte, error) {
	return ioutil.ReadFile(file.Path())
}

// ReadJSON reads the config file as JSON to the given object. Make sure the config file exists!
func (file *File) ReadJSON(object interface{}) error {
	data, err := file.read()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, object)
}

// write writes the given data to the config file (and creates parent directories if necessary).
func (file *File) write(data []byte) error {
	if err := os.MkdirAll(file.dir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(file.Path(), data, 0600)
}

// WriteJSON writes the given object as JSON to the config file.
func (file *File) WriteJSON(object interface{}) error {
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}
	return file.write(data)
}
