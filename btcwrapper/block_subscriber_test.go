package btcwrapper

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"go.uber.org/zap/zaptest"
	"gotest.tools/assert"

	relayerconfig "github.com/gonative-cc/relayer/bitcoinspv/config"
)

func TestSetupBitcoindConnection(t *testing.T) {
	tests := []struct {
		name        string
		client      *Client
		expectError error
	}{
		{
			name: "valid client",
			client: &Client{
				Client:                nil,
				zeromqClient:          nil,
				chainParams:           &chaincfg.SimNetParams,
				config:                &relayerconfig.DefaultConfig().BTC,
				logger:                zaptest.NewLogger(t).Sugar(),
				blockEventsChannel:    nil,
				retrySleepDuration:    time.Second,
				maxRetrySleepDuration: time.Minute,
			},
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setupBitcoindConnection(tt.client)

			if tt.expectError != nil {
				assert.Error(t, err, tt.expectError.Error())
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, tt.client.Client != nil)
			assert.Assert(t, tt.client.zeromqClient != nil)
		})
	}
}
