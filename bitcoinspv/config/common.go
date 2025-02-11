package config

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

const (
	defaultRetrySleepDuration    = 5 * time.Second
	defaultMaxRetrySleepDuration = 5 * time.Minute
)

// CommonConfig defines the server's basic configuration
type CommonConfig struct {
	// Format is the format of the log (json|auto|console|logfmt)
	Format string `mapstructure:"log-format"`
	// Level is the log level (debug|warn|error|panic|fatal)
	Level string `mapstructure:"log-level"`
	// SleepDuration is the backoff interval for the first retry
	SleepDuration time.Duration `mapstructure:"retry-sleep-duration"`
	// MaxSleepDuration is the maximum backoff interval between retries
	MaxSleepDuration time.Duration `mapstructure:"max-retry-sleep-duration"`
}

func isOneOf(v string, list []string) bool {
	for _, item := range list {
		if v == item {
			return true
		}
	}
	return false
}

// Validate does validation checks for common configration values
func (cfg *CommonConfig) Validate() error {
	if !isOneOf(cfg.Format, []string{"json", "auto", "console", "logfmt"}) {
		return errors.New("log-format is not one of json|auto|console|logfmt")
	}
	if !isOneOf(cfg.Level, []string{"debug", "warn", "error", "panic", "fatal"}) {
		return errors.New("log-level is not one of debug|warn|error|panic|fatal")
	}
	if cfg.SleepDuration < 0 {
		return errors.New("retry-sleep-duration can't be negative")
	}
	if cfg.MaxSleepDuration < 0 {
		return errors.New("max-retry-sleep-duration can't be negative")
	}
	return nil
}

// CreateLogger creates and returns root logger
func (cfg *CommonConfig) CreateLogger() (*zap.Logger, error) {
	return NewRootLogger(cfg.Format, cfg.Level)
}

// DefaultCommonConfig returns default values for common config
func DefaultCommonConfig() CommonConfig {
	return CommonConfig{
		Format:           "auto",
		Level:            "debug",
		SleepDuration:    defaultRetrySleepDuration,
		MaxSleepDuration: defaultMaxRetrySleepDuration,
	}
}
