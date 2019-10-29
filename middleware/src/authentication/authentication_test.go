package authentication_test

import (
	"testing"

	authentication "github.com/digitalbitbox/bitbox-base/middleware/src/authentication"
	"github.com/stretchr/testify/require"
)

func TestAuthentication(t *testing.T) {
	testAuthentication, err := authentication.NewJwtAuth()
	require.NoError(t, err)
	token, err := testAuthentication.GenerateToken("admin")
	require.NoError(t, err)
	err = testAuthentication.ValidateToken(token)
	t.Log(token)
	require.NoError(t, err)
	err = testAuthentication.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNTcwODgyNTM0fQ.iV2cdaWB4GQ7Ux7jRFg0smJZWvDtOEUcndMDJFtZUoQ")
	require.Error(t, err, "signature is invalid")
}
