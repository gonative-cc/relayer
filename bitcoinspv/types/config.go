package types

import (
	"fmt"
)

const (
	minBTCCacheSize = 1000
	maxHeadersInMsg = 100 // maximum number of headers in a MsgInsertHeaders message
)

func GetValidNetParams() map[string]bool {
	params := map[string]bool{
		BtcMainnet.String(): true,
		BtcTestnet.String(): true,
		BtcSimnet.String():  true,
		BtcRegtest.String(): true,
		BtcSignet.String():  true,
	}

	return params
}

func GetValidBtcBackends() map[SupportedBtcBackend]bool {
	validBtcBackends := map[SupportedBtcBackend]bool{
		Bitcoind: true,
		Btcd:     true,
	}

	return validBtcBackends
}

// RelayerConfig defines configuration for the relayer.
type RelayerConfig struct {
	NetParams       string `mapstructure:"netparams"`          // should be mainnet|testnet|simnet|signet
	BTCCacheSize    uint64 `mapstructure:"btc_cache_size"`     // size of the BTC cache
	MaxHeadersInMsg uint32 `mapstructure:"max_headers_in_msg"` // maximum number of headers in a MsgInsertHeaders message
}

func (cfg *RelayerConfig) Validate() error {
	if _, ok := GetValidNetParams()[cfg.NetParams]; !ok {
		return fmt.Errorf("invalid net params")
	}
	if cfg.BTCCacheSize < minBTCCacheSize {
		return fmt.Errorf("BTC cache size has to be at least %d", minBTCCacheSize)
	}
	if cfg.MaxHeadersInMsg < maxHeadersInMsg {
		return fmt.Errorf("max_headers_in_msg has to be at least %d", maxHeadersInMsg)
	}
	return nil
}

func DefaultRelayerConfig() RelayerConfig {
	return RelayerConfig{
		NetParams:       BtcSimnet.String(),
		BTCCacheSize:    minBTCCacheSize,
		MaxHeadersInMsg: maxHeadersInMsg,
	}
}
