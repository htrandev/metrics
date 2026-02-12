package crypto

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	privateKeyFile = filepath.Join("testdata", "test.rsa")
	publicKeyFile  = filepath.Join("testdata", "test.rsa.pub")
)

func TestRsa(t *testing.T) {
	privateKey, err := PrivateKey(privateKeyFile)
	require.NoError(t, err)
	require.NotNil(t, privateKey)

	publicKey, err := PublicKey(publicKeyFile)
	require.NoError(t, err)
	require.NotNil(t, publicKey)

	testdata := []byte("testdata")

	encrypted, err := Encrypt(publicKey, testdata)
	require.NoError(t, err)

	decrypted, err := Decrypt(privateKey, encrypted)
	require.NoError(t, err)

	require.Equal(t, testdata, decrypted)
}
