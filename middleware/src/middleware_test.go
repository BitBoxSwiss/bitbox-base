package middleware_test

import (
	"testing"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	basemessages "github.com/digitalbitbox/bitbox-base/middleware/src/messages"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	middlewareInstance := middleware.NewMiddleware("user", "password", "8332", "/home/bitcoin/.lightning", "18442", "testnet")
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
}
