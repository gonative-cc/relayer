package cli

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestLoadConfig(t *testing.T) {
	// temporary file with test config
	tempFile, err := os.CreateTemp("", "config_test.yaml")
	assert.NilError(t, err)
	defer os.Remove(tempFile.Name())

	configYAML := `
        bitcoin:
          rpc_host: "test-rpc-host"
          rpc_user: "test-rpc-user"
          rpc_pass: "test-rpc-pass"
          confirmation_threshold: 3
          http_post_mode: true
          disable_tls: false
          network: "regtest"
    `
	_, err = tempFile.WriteString(configYAML)
	assert.NilError(t, err)

	config, err := loadConfig(tempFile.Name())
	assert.NilError(t, err)
	assert.Equal(t, "test-rpc-host", config.Btc.RPCHost)
	assert.Equal(t, "test-rpc-user", config.Btc.RPCUser)
	assert.Equal(t, "test-rpc-pass", config.Btc.RPCPass)
	assert.Equal(t, uint8(3), config.Btc.ConfirmationThreshold)
	assert.Equal(t, true, config.Btc.HTTPPostMode)
	assert.Equal(t, false, config.Btc.DisableTLS)
	assert.Equal(t, "regtest", config.Btc.Network)
}
