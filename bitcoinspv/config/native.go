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

// Validate does validation checks for native client configration values
func (cfg *NativeConfig) Validate() error {
	if cfg.RPCEndpoint == "" {
		return errors.New("native RPC endpoint cannot be empty")
	}

	return nil
}

// DefaultNativeConfig returns default values for native node config
func DefaultNativeConfig() NativeConfig {
	return NativeConfig{
		RPCEndpoint: DefaultNativeRPCEndpoint,
	}
}
