package types

import (
	"fmt"
)

const (
	minBTCCacheSize = 1000
	maxHeadersInMsg = 100 // maximum number of headers in a MsgInsertHeaders message
)

// GetValidNetParams returns a map of valid Bitcoin network parameters
func GetValidNetParams() map[string]bool {
	return map[string]bool{
		BtcMainnet.String(): true,
		BtcTestnet.String(): true,
		BtcSimnet.String():  true,
		BtcRegtest.String(): true,
		BtcSignet.String():  true,
	}
}

// GetValidBtcBackends returns a map of supported Bitcoin backend types
func GetValidBtcBackends() map[SupportedBtcBackend]bool {
	return map[SupportedBtcBackend]bool{
		Bitcoind: true,
		Btcd:     true,
	}
}

// RelayerConfig defines configuration for the relayer.
type RelayerConfig struct {
	// NetParams should be mainnet|testnet|simnet|signet
	NetParams string `mapstructure:"netparams"`
	// BTCCacheSize is the size of the BTC cache
	BTCCacheSize uint64 `mapstructure:"btc_cache_size"`
	// MaxHeadersInMsg is the maximum number of headers in a MsgInsertHeaders message
	MaxHeadersInMsg uint32 `mapstructure:"max_headers_in_msg"`
}

// Validate checks if the RelayerConfig values are valid
func (cfg *RelayerConfig) Validate() error {
	for _, validate := range []func() error{
		cfg.validateNetParams,
		cfg.validateCacheSize,
		cfg.validateHeadersLimit,
	} {
		if err := validate(); err != nil {
			return err
		}
	}
	return nil
}

func (cfg *RelayerConfig) validateNetParams() error {
	if _, ok := GetValidNetParams()[cfg.NetParams]; !ok {
		return fmt.Errorf("invalid net params")
	}
	return nil
}

func (cfg *RelayerConfig) validateCacheSize() error {
	if cfg.BTCCacheSize < minBTCCacheSize {
		return fmt.Errorf("BTC cache size has to be at least %d", minBTCCacheSize)
	}
	return nil
}

func (cfg *RelayerConfig) validateHeadersLimit() error {
	if cfg.MaxHeadersInMsg < maxHeadersInMsg {
		return fmt.Errorf("max_headers_in_msg has to be at least %d", maxHeadersInMsg)
	}
	return nil
}

// DefaultRelayerConfig returns a RelayerConfig with default values
func DefaultRelayerConfig() RelayerConfig {
	return RelayerConfig{
		NetParams:       BtcSimnet.String(),
		BTCCacheSize:    minBTCCacheSize,
		MaxHeadersInMsg: maxHeadersInMsg,
	}
}
