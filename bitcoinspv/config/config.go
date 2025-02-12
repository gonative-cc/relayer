package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	defaultAppDataDir = btcutil.AppDataDir("native-bitcoin-spv", false)
	defaultConfigFile = filepath.Join(defaultAppDataDir, "bitcoin-spv.yml")
)

// Config defines the server's top level configuration.
type Config struct {
	BTC     BTCConfig     `mapstructure:"btc"`
	Native  NativeConfig  `mapstructure:"native"`
	Relayer RelayerConfig `mapstructure:"relayer"`
}

// Validate validates all the configuration options.
func (cfg *Config) Validate() error {
	if err := cfg.BTC.Validate(); err != nil {
		return fmt.Errorf("invalid config in btc: %w", err)
	}

	if err := cfg.Native.Validate(); err != nil {
		return fmt.Errorf("invalid config in native: %w", err)
	}

	if err := cfg.Relayer.Validate(); err != nil {
		return fmt.Errorf("invalid config in relayer: %w", err)
	}

	return nil
}

// CreateLogger creates and returns a logger from common config values
func (cfg *Config) CreateLogger() (*zap.Logger, error) {
	return NewRootLogger(cfg.Relayer.Format, cfg.Relayer.Level)
}

// DefaultConfigFile returns the default config file path
func DefaultConfigFile() string {
	return defaultConfigFile
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		BTC:     DefaultBTCConfig(),
		Native:  DefaultNativeConfig(),
		Relayer: DefaultRelayerConfig(),
	}
}

// New returns a fully parsed Config object from a given file directory
func New(configFile string) (Config, error) {
	if _, err := os.Stat(configFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("no config file found at %s", configFile)
		}
		return Config{}, err
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
