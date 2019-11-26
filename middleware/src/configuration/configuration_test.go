package configuration_test

import (
	"testing"

	"github.com/digitalbitbox/bitbox-base/middleware/src/configuration"
	"github.com/stretchr/testify/require"
)

// This tests unit tests the getters and the constructor of the `configuration`
// package.
func TestConfiguration(t *testing.T) {
	const (
		bbbCmdScript              string = "/path/to/cmd-script.sh"
		bbbConfigScript           string = "/path/to/config-script.sh"
		bbbSystemctlScript        string = "/path/to/systemctl-script.sh"
		electrsRPCPort            string = "18442"
		imageUpdateInfoURL        string = "https://shiftcrypto.ch/updates/base.json"
		middlewarePort            string = "8085"
		middlewareVersion         string = "0.0.1"
		network                   string = "testnet"
		notificationNamedPipePath string = "/tmp/middleware-notification.pipe"
		prometheusURL             string = "http://localhost:9090"
		redisMock                 bool   = false
		redisPort                 string = "6379"
	)

	config := configuration.NewConfiguration(
		bbbCmdScript, bbbConfigScript, bbbSystemctlScript, electrsRPCPort,
		imageUpdateInfoURL, middlewarePort, middlewareVersion, network,
		notificationNamedPipePath, prometheusURL, redisMock, redisPort,
	)

	require.Equal(t, bbbCmdScript, config.GetBBBCmdScript())
	require.Equal(t, bbbConfigScript, config.GetBBBConfigScript())
	require.Equal(t, bbbSystemctlScript, config.GetBBBSystemctlScript())
	require.Equal(t, electrsRPCPort, config.GetElectrsRPCPort())
	require.Equal(t, imageUpdateInfoURL, config.GetImageUpdateInfoURL())
	require.Equal(t, middlewarePort, config.GetMiddlewarePort())
	require.Equal(t, middlewareVersion, config.GetMiddlewareVersion())
	require.Equal(t, network, config.GetNetwork())
	require.Equal(t, notificationNamedPipePath, config.GetNotificationNamedPipePath())
	require.Equal(t, prometheusURL, config.GetPrometheusURL())
	require.Equal(t, redisPort, config.GetRedisPort())
}
