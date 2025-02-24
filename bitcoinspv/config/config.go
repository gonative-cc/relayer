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

const (
	appName     = "native-bitcoin-spv"
	cfgFileName = "bitcoin-spv.yml"
)

var (
	defaultCfgDir  = btcutil.AppDataDir(appName, false)
	defaultCfgFile = filepath.Join(defaultCfgDir, cfgFileName)
)

// Config represents the main configuration structure for the application
type Config struct {
	Sui     SuiConfig     `mapstucture:"sui"`
	BTC     BTCConfig     `mapstructure:"btc"`
	Native  NativeConfig  `mapstructure:"native"`
	Relayer RelayerConfig `mapstructure:"relayer"`
}

// Validate checks if the configuration is valid by running validation on all components
func (c *Config) Validate() error {
	validators := []struct {
		validator func() error
		name      string
	}{
		{c.BTC.Validate, "btc"},
		{c.Native.Validate, "native"},
		{c.Relayer.Validate, "relayer"},
	}

	for _, v := range validators {
		if err := v.validator(); err != nil {
			return fmt.Errorf("invalid config in %s: %w", v.name, err)
		}
	}

	return nil
}

// CreateLogger creates a new logger instance using the relayer configuration
func (c *Config) CreateLogger() (*zap.Logger, error) {
	return NewRootLogger(c.Relayer.Format, c.Relayer.Level)
}

// DefaultCfgFile returns the default path to the configuration file
func DefaultCfgFile() string {
	return defaultCfgFile
}

// DefaultConfig returns a new Config instance with default values
func DefaultConfig() *Config {
	return &Config{
		BTC:     DefaultBTCConfig(),
		Native:  DefaultNativeConfig(),
		Relayer: DefaultRelayerConfig(),
	}
}

// New creates a new Config instance from the specified configuration file
func New(cfgFile string) (Config, error) {
	if err := validateConfigFile(cfgFile); err != nil {
		return Config{}, err
	}

	v := viper.New()
	v.SetConfigFile(cfgFile)

	if err := v.ReadInConfig(); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateConfigFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no config file found at %s", path)
		}
		return err
	}
	return nil
}
