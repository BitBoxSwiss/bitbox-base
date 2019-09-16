package system_test

import (
	"testing"

	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	"github.com/stretchr/testify/require"
)

func TestSystem(t *testing.T) {
	argumentMap := make(map[string]string)
	argumentMap["bitcoinRPCUser"] = "user"
	argumentMap["bitcoinRPCPassword"] = "password"
	argumentMap["bitcoinRPCPort"] = "8332"
	argumentMap["lightningRPCPath"] = "/home/bitcoin/.lightning"
	argumentMap["electrsRPCPort"] = "18442"
	argumentMap["network"] = "testnet"
	argumentMap["bbbConfigScript"] = "/home/bitcoin/config-script.sh"
	argumentMap["bbbCmdScript"] = "/home/bitcoin/cmd-script.sh"
	argumentMap["prometheusURL"] = "http://localhost:9090"
	argumentMap["redisPort"] = "6379"

	environmentInstance := system.NewEnvironment(argumentMap)
	require.Equal(t, environmentInstance.GetBitcoinRPCPort(), "8332")
	require.Equal(t, environmentInstance.GetBitcoinRPCUser(), "user")
	require.Equal(t, environmentInstance.GetBitcoinRPCPassword(), "password")
	require.Equal(t, environmentInstance.GetLightningRPCPath(), "/home/bitcoin/.lightning")
	require.Equal(t, environmentInstance.GetBBBConfigScript(), "/home/bitcoin/config-script.sh")
	require.Equal(t, environmentInstance.GetBBBCmdScript(), "/home/bitcoin/cmd-script.sh")
	require.Equal(t, environmentInstance.Network, "testnet")
	require.Equal(t, environmentInstance.ElectrsRPCPort, "18442")
	require.Equal(t, environmentInstance.GetPrometheusURL(), "http://localhost:9090")
	require.Equal(t, "6379", environmentInstance.GetRedisPort())

	//test unhappy path
	argumentMap = make(map[string]string)
	argumentMap["lel"] = "1"
	environmentInstance = system.NewEnvironment(argumentMap)
	require.Equal(t, environmentInstance.GetBitcoinRPCPort(), "")

	argumentMap = make(map[string]string)
	environmentInstance = system.NewEnvironment(argumentMap)
	require.Equal(t, environmentInstance.GetBitcoinRPCPort(), "")
}
