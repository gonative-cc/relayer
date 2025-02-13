package btcwrapper

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

// GetBTCNodeParams extracts and returns the BTC node parameters
func GetBTCNodeParams(net string) (*chaincfg.Params, error) {
	switch net {
	case btctypes.Mainnet.String():
		return &chaincfg.MainNetParams, nil
	case btctypes.Testnet.String():
		return &chaincfg.TestNet3Params, nil
	case btctypes.Simnet.String():
		return &chaincfg.SimNetParams, nil
	case btctypes.Regtest.String():
		return &chaincfg.RegressionNetParams, nil
	case btctypes.Signet.String():
		return &chaincfg.SigNetParams, nil
	}
	return nil, fmt.Errorf(
		"invalid BTC network '%s'. Valid networks are: %s, %s, %s, %s, %s",
		net,
		btctypes.Mainnet.String(),
		btctypes.Testnet.String(),
		btctypes.Simnet.String(),
		btctypes.Regtest.String(),
		btctypes.Signet.String(),
	)
}
