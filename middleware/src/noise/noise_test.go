package noisemanager_test

import (
	"testing"

	noisemanager "github.com/digitalbitbox/bitbox-base/middleware/src/noise"
	"github.com/stretchr/testify/require"
)

func TestNoiseRejected(t *testing.T) {
	noiseInstance := noisemanager.NewNoiseConfig(
		".base",
		func([]byte) (bool, error) { return false, nil },
	)
	response, err := noiseInstance.CheckVerification()
	require.NoError(t, err)
	require.Equal(t, string(response), "\x01")
}

func TestNoise(t *testing.T) {
	noiseInstance := noisemanager.NewNoiseConfig(
		".base",
		func([]byte) (bool, error) { return true, nil },
	)
	response, err := noiseInstance.CheckVerification()
	require.NoError(t, err)
	require.Equal(t, string(response), "\x00")
	msg := noiseInstance.Encrypt([]byte("test"))
	if string(msg) == "" {
		t.Error("did not receive error when encrypting from uninitialized noise")
	}
	_, err = noiseInstance.Decrypt([]byte("test"))
	if err == nil {
		t.Error("did not receive error when decrypting from unitialized noise")
	}
}
