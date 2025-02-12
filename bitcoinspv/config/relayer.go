package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultRetrySleepDuration    = 5 * time.Second
	defaultMaxRetrySleepDuration = 5 * time.Minute
	minBTCCacheSize              = 1000
	maxHeadersInMsg              = 100 // maximum number of headers in a MsgInsertHeaders message
)

// RelayerConfig defines configuration for the spv relayer.
type RelayerConfig struct {
	// Format is the format of the log (json|auto|console|logfmt)
	Format string `mapstructure:"log-format"`
	// Level is the log level (debug|warn|error|panic|fatal)
	Level string `mapstructure:"log-level"`
	// SleepDuration is the backoff interval for the first retry
	SleepDuration time.Duration `mapstructure:"retry-sleep-duration"`
	// MaxSleepDuration is the maximum backoff interval between retries
	MaxSleepDuration time.Duration `mapstructure:"max-retry-sleep-duration"`
	// NetParams should be mainnet|testnet|simnet|signet
	NetParams string `mapstructure:"netparams"`
	// BTCCacheSize is size of the BTC cache
	BTCCacheSize int64 `mapstructure:"btc_cache_size"`
	// MaxHeadersInMsg is maximum number of headers in a MsgInsertHeaders message
	MaxHeadersInMsg uint32 `mapstructure:"max_headers_in_msg"`
}

func isPresent(v string, list []string) bool {
	for _, item := range list {
		if v == item {
			return true
		}
	}
	return false
}

// CreateLogger creates and returns root logger
func (cfg *RelayerConfig) CreateLogger() (*zap.Logger, error) {
	return NewRootLogger(cfg.Format, cfg.Level)
}

// Validate does validation checks for relayer configration values
func (cfg *RelayerConfig) Validate() error {
	if !isPresent(cfg.Format, []string{"json", "auto", "console", "logfmt"}) {
		return errors.New("log-format is not one of json|auto|console|logfmt")
	}
	if !isPresent(cfg.Level, []string{"debug", "warn", "error", "panic", "fatal"}) {
		return errors.New("log-level is not one of debug|warn|error|panic|fatal")
	}
	if cfg.SleepDuration < 0 {
		return errors.New("retry-sleep-duration can't be negative")
	}
	if cfg.MaxSleepDuration < 0 {
		return errors.New("max-retry-sleep-duration can't be negative")
	}
	if err := cfg.validateNetParams(); err != nil {
		return err
	}
	if err := cfg.validateBTCCacheSize(); err != nil {
		return err
	}
	err := cfg.validateMaxHeadersInMsg()
	return err
}

func (cfg *RelayerConfig) validateNetParams() error {
	if _, ok := types.GetValidNetParams()[cfg.NetParams]; !ok {
		return fmt.Errorf("invalid net params: %s", cfg.NetParams)
	}
	return nil
}

func (cfg *RelayerConfig) validateBTCCacheSize() error {
	if cfg.BTCCacheSize < minBTCCacheSize {
		return fmt.Errorf("BTC cache size has to be at least %d", minBTCCacheSize)
	}
	return nil
}

func (cfg *RelayerConfig) validateMaxHeadersInMsg() error {
	if cfg.MaxHeadersInMsg < maxHeadersInMsg {
		return fmt.Errorf("max_headers_in_msg has to be at least %d", maxHeadersInMsg)
	}
	return nil
}

// DefaultRelayerConfig returns default values for relayer config
func DefaultRelayerConfig() RelayerConfig {
	return RelayerConfig{
		Format:           "auto",
		Level:            "debug",
		SleepDuration:    defaultRetrySleepDuration,
		MaxSleepDuration: defaultMaxRetrySleepDuration,
		NetParams:        types.BtcTestnet.String(),
		BTCCacheSize:     minBTCCacheSize,
		MaxHeadersInMsg:  maxHeadersInMsg,
	}
}

// NewRootLogger creates a new logger object with the given format and log level
// (copied from https://github.com/cosmos/relayer/blob/v2.4.2/cmd/root.go#L174-L202)
func NewRootLogger(format string, logLevel string) (*zap.Logger, error) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format("2006-01-02T15:04:05.000000Z07:00"))
	}
	config.LevelKey = "lvl"

	var enc zapcore.Encoder
	switch format {
	case "json":
		enc = zapcore.NewJSONEncoder(config)
	case "auto", "console":
		enc = zapcore.NewConsoleEncoder(config)
	case "logfmt":
		enc = zaplogfmt.NewEncoder(config)
	default:
		return nil, fmt.Errorf("unrecognized log format %q", format)
	}

	level := zapcore.InfoLevel
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	}
	return zap.New(zapcore.NewCore(
		enc,
		os.Stdout,
		level,
	)), nil
}
