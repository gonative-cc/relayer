package config

import "time"

// BabylonConfig defines configuration for the Babylon client
// adapted from https://github.com/strangelove-ventures/lens/blob/v0.5.1/client/config.go
type BabylonConfig struct {
	Key              string        `mapstructure:"key"`
	ChainID          string        `mapstructure:"chain-id"`
	RPCAddr          string        `mapstructure:"rpc-addr"`
	GRPCAddr         string        `mapstructure:"grpc-addr"`
	AccountPrefix    string        `mapstructure:"account-prefix"`
	KeyringBackend   string        `mapstructure:"keyring-backend"`
	GasAdjustment    float64       `mapstructure:"gas-adjustment"`
	GasPrices        string        `mapstructure:"gas-prices"`
	KeyDirectory     string        `mapstructure:"key-directory"`
	Debug            bool          `mapstructure:"debug"`
	Timeout          time.Duration `mapstructure:"timeout"`
	BlockTimeout     time.Duration `mapstructure:"block-timeout"`
	OutputFormat     string        `mapstructure:"output-format"`
	SignModeStr      string        `mapstructure:"sign-mode"`
	SubmitterAddress string        `mapstructure:"submitter-address"`
}
