package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetServerConfig(t *testing.T) {
	os.Setenv("CONFIG","")
	_, err := GetServerConfig()
	require.NoError(t, err)
}
