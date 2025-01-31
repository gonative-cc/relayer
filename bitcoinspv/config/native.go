package config

import (
	"errors"
)

const (
	DefaultNativeRPCEndpoint = "localhost:9797"
	errEmptyRPCEndpoint      = "native RPC endpoint cannot be empty"
)

// NativeConfig defines the native RPC server configuration
type NativeConfig struct {
	RPCEndpoint string `mapstructure:"rpc-endpoint"`
}

func (cfg *NativeConfig) Validate() error {
	if cfg.RPCEndpoint == "" {
		return errors.New(errEmptyRPCEndpoint)
	}

	return nil
}

func DefaultNativeConfig() NativeConfig {
	return NativeConfig{
		RPCEndpoint: DefaultNativeRPCEndpoint,
	}
}
