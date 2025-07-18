package config

import (
	"errors"
)

const (
	defaultNativeRPCEndpoint = "http://localhost:9797"
)

// NativeConfig defines the native RPC server configuration
type NativeConfig struct {
	RPCEndpoint string `mapstructure:"rpc-endpoint"`
}

// Validate does validation checks for native client configuration values
func (cfg *NativeConfig) Validate() error {
	if cfg.RPCEndpoint == "" {
		return errors.New("native RPC endpoint cannot be empty")
	}

	return nil
}

// DefaultNativeConfig returns default values for native node config
func DefaultNativeConfig() NativeConfig {
	return NativeConfig{
		RPCEndpoint: defaultNativeRPCEndpoint,
	}
}
