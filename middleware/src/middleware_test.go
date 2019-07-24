package middleware_test

import (
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	basemessages "github.com/digitalbitbox/bitbox-base/middleware/src/messages"
	"github.com/golang/protobuf/proto"
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
	marshalled := middlewareInstance.SystemEnv()
	unmarshalled := &basemessages.BitBoxBaseOut{}
	err := proto.Unmarshal(marshalled, unmarshalled)
	require.NoError(t, err)

	unmarshalledSystemEnv, ok := unmarshalled.BitBoxBaseOut.(*basemessages.BitBoxBaseOut_BaseSystemEnvOut)
	if !ok {
		t.Error("Protobuf parsing into system env message failed")
	}
	port := unmarshalledSystemEnv.BaseSystemEnvOut.ElectrsRPCPort
	require.Equal(t, port, "18442")
	network := unmarshalledSystemEnv.BaseSystemEnvOut.Network
	require.Equal(t, network, "testnet")

	marshalled = middlewareInstance.ResyncBitcoin()
	err = proto.Unmarshal(marshalled, unmarshalled)
	require.NoError(t, err)
	_, ok = unmarshalled.BitBoxBaseOut.(*basemessages.BitBoxBaseOut_BaseResyncOut)
	if !ok {
		t.Error("Protobuf parsing into resync bitcoin message failed")
	}
}
