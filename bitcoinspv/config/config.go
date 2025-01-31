package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const defaultConfigFilename = "bitcoin-spv.yml"

var (
	defaultBtcCAFile       = filepath.Join(btcutil.AppDataDir("btcd", false), "rpc.cert")
	defaultBtcWalletCAFile = filepath.Join(btcutil.AppDataDir("btcwallet", false), "rpc.cert")
	defaultAppDataDir      = btcutil.AppDataDir("native-bitcoin-spv", false)
	defaultConfigFile      = filepath.Join(defaultAppDataDir, defaultConfigFilename)
)

// Config defines the server's top level configuration
type Config struct {
	Common CommonConfig `mapstructure:"common"`
	BTC    BTCConfig    `mapstructure:"btc"`
	Native NativeConfig `mapstructure:"native"`
	// Metrics MetricsConfig `mapstructure:"metrics"`
	Relayer RelayerConfig `mapstructure:"relayer"`
}

func (cfg *Config) Validate() error {
	if err := cfg.Common.Validate(); err != nil {
		return fmt.Errorf("invalid config in common: %w", err)
	}

	if err := cfg.BTC.Validate(); err != nil {
		return fmt.Errorf("invalid config in btc: %w", err)
	}

	if err := cfg.Native.Validate(); err != nil {
		return fmt.Errorf("invalid config in native: %w", err)
	}

	// if err := cfg.Metrics.Validate(); err != nil {
	// 	return fmt.Errorf("invalid config in metrics: %w", err)
	// }

	if err := cfg.Relayer.Validate(); err != nil {
		return fmt.Errorf("invalid config in relayer: %w", err)
	}

	return nil
}

func (cfg *Config) CreateLogger() (*zap.Logger, error) {
	return cfg.Common.CreateLogger()
}

func DefaultConfigFile() string {
	return defaultConfigFile
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Common: DefaultCommonConfig(),
		BTC:    DefaultBTCConfig(),
		Native: DefaultNativeConfig(),
		// Metrics: DefaultMetricsConfig(),
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

func WriteSample() error {
	cfg := DefaultConfig()
	d, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	// write to file
	return os.WriteFile("./sample-bitcoin-spv.yml", d, 0600)
}
