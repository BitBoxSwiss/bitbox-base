package hsm

import (
	"log"

	"github.com/digitalbitbox/bitbox02-api-go/communication/usart"
	"github.com/flynn/noise"
)

// See ConfigInterace: https://github.com/digitalbitbox/bitbox02-api-go/blob/e8ae46debc009cfc7a64f45ec191de0220f0c401/api/firmware/device.go#L50
type bitbox02Config struct{}

// ContainsDeviceStaticPubkey implements firmware.ConfigInterface
func (bb02Config *bitbox02Config) ContainsDeviceStaticPubkey(pubkey []byte) bool {
	return false
}

// AddDeviceStaticPubkey implements firmware.ConfigInterface
func (bb02Config *bitbox02Config) AddDeviceStaticPubkey(pubkey []byte) error {
	return nil
}

// GetAppNoiseStaticKeypair implements firmware.ConfigInterface
func (bb02Config *bitbox02Config) GetAppNoiseStaticKeypair() *noise.DHKey {
	return nil
}

// SetAppNoiseStaticKeypair implements firmware.ConfigInterface
func (bb02Config *bitbox02Config) SetAppNoiseStaticKeypair(key *noise.DHKey) error {
	return nil
}

type bitbox02Logger struct{}

// Error implements firmware.Logger
func (bb02Logger *bitbox02Logger) Error(msg string, err error) {
	log.Println(msg, err)
}

// Info implements firmware.Logger
func (bb02Logger *bitbox02Logger) Info(msg string) {
	log.Println(msg)
}

// Debug implements firmware.Logger
func (bb02Logger *bitbox02Logger) Debug(msg string) {
	log.Println(msg)
}

// just translating SendFrame with incompatible signature (string<->[]byte), will be made consistent
// later...
type usartCommunication struct {
	*usart.Communication
}

// SendFrame implements firmware.Communication.
func (communication usartCommunication) SendFrame(msg string) error {
	return communication.Communication.SendFrame([]byte(msg))
}
