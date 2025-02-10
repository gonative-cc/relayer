package config

import (
	"errors"
)

const (
	DefaultNativeRPCEndpoint = "http://localhost:9797"
)

// NativeConfig defines the native RPC server configuration
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
