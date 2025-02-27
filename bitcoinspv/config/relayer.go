package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultRetrySleepDuration    = 5 * time.Second
	defaultMaxRetrySleepDuration = 5 * time.Minute
	minBTCCacheSize              = 1000
	minheadersChunkSize          = 1
)

// RelayerConfig defines configuration for the spv relayer.
type RelayerConfig struct {
	// Format is the format of the log (json|auto|console|logfmt)
	Format string `mapstructure:"log-format"`
	// Level is the log level (debug|warn|error|panic|fatal)
	Level string `mapstructure:"log-level"`
	// NetParams should be mainnet|testnet|simnet|signet
	NetParams string `mapstructure:"netparams"`
	// SleepDuration is the backoff interval for the first retry
	SleepDuration time.Duration `mapstructure:"retry-sleep-duration"`
	// MaxSleepDuration is the maximum backoff interval between retries
	MaxSleepDuration time.Duration `mapstructure:"max-retry-sleep-duration"`
	// BTCCacheSize is size of the BTC cache
	BTCCacheSize int64 `mapstructure:"cache-size"`
	// HeadersChunkSize is maximum number of headers in a MsgInsertHeaders message
	HeadersChunkSize uint32 `mapstructure:"headers-chunk-size"`
	// ProcessBlockTimeout is the timeout duration for processing a single block.
	ProcessBlockTimeout time.Duration `mapstructure:"process-block-timeout"`
}

func isPresent(v string, list []string) bool {
	for i := range list {
		if list[i] == v {
			return true
		}
	}
	return false
}

// Validate does validation checks for relayer configration values
func (cfg *RelayerConfig) Validate() error {
	if err := cfg.validateLogging(); err != nil {
		return err
	}
	if err := cfg.validateSleepDurations(); err != nil {
		return err
	}
	if err := cfg.validateNetParams(); err != nil {
		return err
	}
	if err := cfg.validateBTCCacheSize(); err != nil {
		return err
	}
	err := cfg.validateHeadersChunkSize()
	return err
}

func (cfg *RelayerConfig) validateLogging() error {
	if !isPresent(cfg.Format, []string{"json", "auto", "console", "logfmt"}) {
		return errors.New("log-format is not one of json|auto|console|logfmt")
	}
	if !isPresent(cfg.Level, []string{"debug", "warn", "error", "panic", "fatal"}) {
		return errors.New("log-level is not one of debug|warn|error|panic|fatal")
	}
	return nil
}

func (cfg *RelayerConfig) validateSleepDurations() error {
	if cfg.SleepDuration < 0 {
		return errors.New("retry-sleep-duration can't be negative")
	}
	if cfg.MaxSleepDuration < 0 {
		return errors.New("max-retry-sleep-duration can't be negative")
	}
	return nil
}

func (cfg *RelayerConfig) validateNetParams() error {
	if _, ok := btctypes.GetValidNetParams()[cfg.NetParams]; !ok {
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

func (cfg *RelayerConfig) validateHeadersChunkSize() error {
	if cfg.HeadersChunkSize < minheadersChunkSize {
		return fmt.Errorf("headers-chunk-size has to be at least %d", minheadersChunkSize)
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
		NetParams:        btctypes.Testnet.String(),
		BTCCacheSize:     minBTCCacheSize,
		HeadersChunkSize: minheadersChunkSize,
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
