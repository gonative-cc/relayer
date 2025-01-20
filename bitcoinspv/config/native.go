package config

import (
	"errors"
)

const (
	DefaultNativeRPCEndpoint = "localhost:9797"
)

// CommonConfig defines the server's basic configuration
type NativeConfig struct {
	RPCEndpoint string `mapstructure:"rpc-endpoint"`
}

func (cfg *NativeConfig) Validate() error {
	if cfg.RPCEndpoint == "" {
		return errors.New("native RPC endpoint cannot be empty")
	}

	return nil
}

func DefaultNativeConfig() NativeConfig {
	return NativeConfig{
		RPCEndpoint: DefaultNativeRPCEndpoint,
	}
}
