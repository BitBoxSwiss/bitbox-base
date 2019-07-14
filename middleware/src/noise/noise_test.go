package noisemanager_test

import (
	"testing"

	noisemanager "github.com/digitalbitbox/bitbox-base/middleware/src/noise"
	"github.com/stretchr/testify/require"
)

func TestNoise(t *testing.T) {
	noiseInstance := noisemanager.NewNoiseConfig()
	response := noiseInstance.CheckVerification()
	require.Equal(t, string(response), "\x00")
	_, err := noiseInstance.Encrypt([]byte("test"))
	if err == nil {
		t.Error("did not receive error when encrypting from uninitialized noise")
	}
	_, err = noiseInstance.Decrypt([]byte("test"))
	if err == nil {
		t.Error("did not receive error when decrypting from unitialized noise")
	}
}
