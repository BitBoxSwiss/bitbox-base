package system_test

import (
	"testing"

	"github.com/digitalbitbox/bitbox-base/middleware/src/system"
	"github.com/stretchr/testify/require"
)

func TestSystem(t *testing.T) {
	environmentInstance := system.NewEnvironment("user", "password", "8332", "/home/bitcoin/.lightning", "18442", "testnet")
	require.Equal(t, environmentInstance.GetBitcoinRPCPort(), "8332")
	require.Equal(t, environmentInstance.GetBitcoinRPCUser(), "user")
	require.Equal(t, environmentInstance.GetBitcoinRPCPassword(), "password")
	require.Equal(t, environmentInstance.GetLightningRPCPath(), "/home/bitcoin/.lightning")
	require.Equal(t, environmentInstance.Network, "testnet")
	require.Equal(t, environmentInstance.ElectrsRPCPort, "18442")
}
