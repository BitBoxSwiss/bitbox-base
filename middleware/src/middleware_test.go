package middleware_test

import (
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	argumentMap := make(map[string]string)
	argumentMap["bitcoinRPCUser"] = "user"
	argumentMap["bitcoinRPCPassword"] = "password"
	argumentMap["bitcoinRPCPort"] = "8332"
	argumentMap["lightningRPCPath"] = "/home/bitcoin/.lightning"
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"
	argumentMap["bbbConfigScript"] = "/home/bitcoin/script.sh"

	middlewareInstance := middleware.NewMiddleware(argumentMap)

	systemEnvResponse, err := middlewareInstance.SystemEnv()
	require.NoError(t, err)
	require.Equal(t, systemEnvResponse.ElectrsRPCPort, "18442")
	require.Equal(t, systemEnvResponse.Network, "testnet")
	resyncBitcoinResponse, err := middlewareInstance.ResyncBitcoin()
	require.NoError(t, err)
	require.Equal(t, resyncBitcoinResponse.Success, false)
}
