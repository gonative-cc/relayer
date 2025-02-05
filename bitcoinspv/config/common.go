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

// LogConfig contains logging related configuration
type LogConfig struct {
	// Format is the format of the log (json|auto|console|logfmt)
	Format string `mapstructure:"log-format"`
	// Level is the log level (debug|warn|error|panic|fatal)
	Level string `mapstructure:"log-level"`
}

// RetryConfig contains retry related configuration
type RetryConfig struct {
	// SleepDuration is the backoff interval for the first retry
	SleepDuration time.Duration `mapstructure:"retry-sleep-duration"`
	// MaxSleepDuration is the maximum backoff interval between retries
	MaxSleepDuration time.Duration `mapstructure:"max-retry-sleep-duration"`
}

// CommonConfig defines the server's basic configuration
type CommonConfig struct {
	LogConfig
	RetryConfig
}

func isOneOf(v string, list []string) bool {
	for _, item := range list {
		if v == item {
			return true
		}
	}
	return false
}

func (cfg *CommonConfig) Validate() error {
	if !isOneOf(cfg.Format, []string{"json", "auto", "console", "logfmt"}) {
		return errors.New("log-format is not one of json|auto|console|logfmt")
	}
	if !isOneOf(cfg.Level, []string{"debug", "warn", "error", "panic", "fatal"}) {
		return errors.New("log-level is not one of debug|warn|error|panic|fatal")
	}
	if cfg.SleepDuration < 0 {
		return errors.New("retry-sleep-time can't be negative")
	}
	if cfg.MaxSleepDuration < 0 {
		return errors.New("max-retry-sleep-time can't be negative")
	}
	return nil
}

func (cfg *CommonConfig) CreateLogger() (*zap.Logger, error) {
	return NewRootLogger(cfg.Format, cfg.Level)
}

func DefaultCommonConfig() CommonConfig {
	return CommonConfig{
		LogConfig: LogConfig{
			Format: "auto",
			Level:  "debug",
		},
		RetryConfig: RetryConfig{
			SleepDuration:    defaultRetrySleepDuration,
			MaxSleepDuration: defaultMaxRetrySleepDuration,
		},
	}
}
