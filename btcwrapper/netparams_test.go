package btcwrapper

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"gotest.tools/assert"
)

func TestGetBTCNodeParams(t *testing.T) {
	tests := []struct {
		name           string
		network        string
		expectedParams *chaincfg.Params
		expectError    bool
	}{
		{
			name:           "mainnet",
			network:        "mainnet",
			expectedParams: &chaincfg.MainNetParams,
			expectError:    false,
		},
		{
			name:           "regtest",
			network:        "regtest",
			expectedParams: &chaincfg.RegressionNetParams,
			expectError:    false,
		},
		{
			name:        "invalid network",
			network:     "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := GetBTCNodeParams(tt.network)

			if tt.expectError {
				assert.Error(
					t, err,
					"invalid BTC network '%s'. Valid networks are: mainnet, testnet, simnet, regtest, signet", tt.network,
				)
				return
			}

			assert.NilError(t, err)
			assert.Equal(
				t, params, tt.expectedParams,
				"expected Network %v, got %v", tt.expectedParams.Name, params.Name,
			)
		})
	}
}
