package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

// Validate checks if the configuration file is valid by running validation on all components
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
func (c *Config) CreateLogger() (zerolog.Logger, error) {
	format := c.Relayer.Format
	logLevel := c.Relayer.Level

	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		log.Error().Err(err).Str("level", logLevel).Msg("Invalid log level encountered after validation")
		return zerolog.Nop(), fmt.Errorf("invalid log level %q: %w", logLevel, err)
	}

	var writer io.Writer
	outputTarget := os.Stderr

	switch strings.ToLower(format) {
	case "json":
		writer = outputTarget
	case "console":
		writer = zerolog.ConsoleWriter{
			Out:        outputTarget,
			TimeFormat: "15:04:05",
		}
	case "auto":
		if isatty.IsTerminal(outputTarget.Fd()) || isatty.IsCygwinTerminal(outputTarget.Fd()) {
			writer = zerolog.ConsoleWriter{
				Out:        outputTarget,
				TimeFormat: "15:04:05",
			}
		} else {
			writer = outputTarget
		}
	default:
		log.Error().Str("format", format).Msg("Unrecognized log format requested after validation")
		return zerolog.Nop(), fmt.Errorf("unrecognized log format: %q", format)
	}

	logger := zerolog.New(writer).Level(level).With().Timestamp().Logger()

	if level <= zerolog.DebugLevel {
		logger = logger.With().Caller().Logger()
	}

	logger.Info().Str("format", format).Str("level", level.String()).Msg("Logger instance created")

	return logger, nil
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
	if err := validateConfigFileExists(cfgFile); err != nil {
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

func validateConfigFileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no config file found at %s", path)
		}
		return err
	}
	return nil
}
