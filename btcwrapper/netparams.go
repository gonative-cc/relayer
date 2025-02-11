package btcwrapper

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// GetBTCNodeParams extracts and returns the BTC node parameters
func GetBTCNodeParams(net string) (*chaincfg.Params, error) {
	switch net {
	case types.BtcMainnet.String():
		return &chaincfg.MainNetParams, nil
	case types.BtcTestnet.String():
		return &chaincfg.TestNet3Params, nil
	case types.BtcSimnet.String():
		return &chaincfg.SimNetParams, nil
	case types.BtcRegtest.String():
		return &chaincfg.RegressionNetParams, nil
	case types.BtcSignet.String():
		return &chaincfg.SigNetParams, nil
	}
	return nil, fmt.Errorf(
		"BTC network with name %s does not exist. should be one of {%s, %s, %s, %s, %s}",
		net,
		types.BtcMainnet.String(),
		types.BtcTestnet.String(),
		types.BtcSimnet.String(),
		types.BtcRegtest.String(),
		types.BtcSignet.String(),
	)
}
