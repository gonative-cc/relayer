package config

import (
	"errors"
	"fmt"
	"time"

	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

const (
	defaultRetrySleepDuration    = 5 * time.Second
	defaultMaxRetrySleepDuration = 5 * time.Minute
	minBTCCacheSize              = 1000
	minheadersChunkSize          = 1
	defaultConfirmationDepth     = 6
)

// RelayerConfig defines configuration for the spv relayer.
type RelayerConfig struct {
	// Format is the format of the log (json|auto|console|logfmt)
	Format string `mapstructure:"log-format"`
	// Level is the log level (debug|warn|error|panic|fatal)
	Level string `mapstructure:"log-level"`
	// NetParams should be mainnet|testnet|simnet|signet
	NetParams string `mapstructure:"netparams"`
	// RetrySleepDuration is the backoff interval for the first retry
	RetrySleepDuration time.Duration `mapstructure:"retry-sleep-duration"`
	// MaxRetrySleepDuration is the maximum backoff interval between retries
	MaxRetrySleepDuration time.Duration `mapstructure:"max-retry-sleep-duration"`
	// BTCCacheSize is size of the BTC cache
	BTCCacheSize int64 `mapstructure:"cache-size"`
	// BTCConfirmationDepth is the number of recent block headers the
	// relayer keeps in its cache and attempts to re-send to the light client.
	BTCConfirmationDepth int64 `mapstructure:"confirmation_depth"`
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
	if err := cfg.validateBTCConfirmationDepth(); err != nil {
		return err
	}
	err := cfg.validateHeadersChunkSize()
	return err
}

func (cfg *RelayerConfig) validateLogging() error {
	validFormats := []string{"json", "auto", "console"}
	if !isPresent(cfg.Format, validFormats) {
		return fmt.Errorf("log-format %q is not one of %v", cfg.Format, validFormats)
	}
	validLevels := []string{"debug", "warn", "error", "panic", "fatal"}
	if !isPresent(cfg.Level, validLevels) {
		return fmt.Errorf("log-level %q is not one of %v", cfg.Level, validLevels)
	}
	return nil
}

func (cfg *RelayerConfig) validateSleepDurations() error {
	if cfg.RetrySleepDuration < 0 {
		return errors.New("retry-sleep-duration can't be negative")
	}
	if cfg.MaxRetrySleepDuration < 0 {
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

func (cfg *RelayerConfig) validateBTCConfirmationDepth() error {
	if cfg.BTCConfirmationDepth < 1 {
		return fmt.Errorf("BTC confirmation depth must be at least 1")
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
		Format:                "auto",
		Level:                 "debug",
		RetrySleepDuration:    defaultRetrySleepDuration,
		MaxRetrySleepDuration: defaultMaxRetrySleepDuration,
		NetParams:             btctypes.Testnet.String(),
		BTCCacheSize:          minBTCCacheSize,
		HeadersChunkSize:      minheadersChunkSize,
		BTCConfirmationDepth:  defaultConfirmationDepth,
	}
}
