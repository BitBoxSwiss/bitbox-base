package middleware_test

import (
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/rpcmessages"
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
	argumentMap["bbbConfigScript"] = "/home/bitcoin/config-script.sh"
	argumentMap["bbbCmdScript"] = "/home/bitcoin/cmd-script.sh"

	middlewareInstance := middleware.NewMiddleware(argumentMap)

	systemEnvResponse := middlewareInstance.SystemEnv()
	require.Equal(t, systemEnvResponse.ElectrsRPCPort, "18442")
	require.Equal(t, systemEnvResponse.Network, "testnet")
	resyncBitcoinResponse, err := middlewareInstance.ResyncBitcoin(rpcmessages.Resync)
	require.Equal(t, resyncBitcoinResponse.Success, false)
	require.NoError(t, err)
	sampleInfo := middlewareInstance.SampleInfo()
	emptySampleInfo := rpcmessages.SampleInfoResponse{
		Blocks:         0,
		Difficulty:     0.0,
		LightningAlias: "disconnected",
	}
	require.Equal(t, sampleInfo, emptySampleInfo)
	verificationProgress := middlewareInstance.VerificationProgress()
	emptyVerificationProgress := rpcmessages.VerificationProgressResponse{
		Blocks:               0,
		Headers:              0,
		VerificationProgress: 0.0,
	}
	require.Equal(t, verificationProgress, emptyVerificationProgress)
}
