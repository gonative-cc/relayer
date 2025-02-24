package btcwrapper

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"gotest.tools/assert"
)

func TestGetBTCNodeParams(t *testing.T) {
	tests := []struct {
		name           string
		network        string
		expectedParams *chaincfg.Params
		expectError    error
	}{
		{
			name:           "mainnet",
			network:        "mainnet",
			expectedParams: &chaincfg.MainNetParams,
			expectError:    nil,
		},
		{
			name:           "regtest",
			network:        "regtest",
			expectedParams: &chaincfg.RegressionNetParams,
			expectError:    nil,
		},
		{
			name:    "invalid network",
			network: "invalid",
			expectError: fmt.Errorf(
				"invalid BTC network 'invalid'. Valid networks are: mainnet, testnet, simnet, regtest, signet",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := GetBTCNodeParams(tt.network)

			if tt.expectError != nil {
				assert.Error(t, err, tt.expectError.Error())
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
