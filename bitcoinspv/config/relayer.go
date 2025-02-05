package config

import (
	"fmt"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

const (
	minBTCCacheSize = 1000
	maxHeadersInMsg = 100 // maximum number of headers in a MsgInsertHeaders message
)

// RelayerConfig defines configuration for the spv relayer.
type RelayerConfig struct {
	NetParams       string `mapstructure:"netparams"`          // should be mainnet|testnet|simnet|signet
	BTCCacheSize    int64  `mapstructure:"btc_cache_size"`     // size of the BTC cache
	MaxHeadersInMsg uint32 `mapstructure:"max_headers_in_msg"` // maximum number of headers in a MsgInsertHeaders message
}

func (cfg *RelayerConfig) Validate() error {
	if err := cfg.validateNetParams(); err != nil {
		return err
	}
	if err := cfg.validateBTCCacheSize(); err != nil {
		return err
	}
	if err := cfg.validateMaxHeadersInMsg(); err != nil {
		return err
	}
	return nil
}

func (cfg *RelayerConfig) validateNetParams() error {
	if _, ok := types.GetValidNetParams()[cfg.NetParams]; !ok {
		return fmt.Errorf("invalid net params")
	}
	return nil
}

func (cfg *RelayerConfig) validateBTCCacheSize() error {
	if cfg.BTCCacheSize < minBTCCacheSize {
		return fmt.Errorf("BTC cache size has to be at least %d", minBTCCacheSize)
	}
	return nil
}

func (cfg *RelayerConfig) validateMaxHeadersInMsg() error {
	if cfg.MaxHeadersInMsg < maxHeadersInMsg {
		return fmt.Errorf("max_headers_in_msg has to be at least %d", maxHeadersInMsg)
	}
	return nil
}

func DefaultRelayerConfig() RelayerConfig {
	return RelayerConfig{
		NetParams:       types.BtcSimnet.String(),
		BTCCacheSize:    minBTCCacheSize,
		MaxHeadersInMsg: maxHeadersInMsg,
	}
}
