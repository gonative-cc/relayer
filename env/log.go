package env

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// ENV variables
const (
	EnvLogLevel = "LOG_LEVEL"
)

// InitLogger setups zerolog global logger
func InitLogger(lvl zerolog.Level) {
	zerolog.SetGlobalLevel(lvl)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// reads log level from env. Returns NoLevel if not specified.
func getLogLevel() (zerolog.Level, error) {
	logLevelStr := os.Getenv(EnvLogLevel)
	if logLevelStr == "" {
		return zerolog.NoLevel, nil
	}
	lvl, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		return lvl, fmt.Errorf("invalid LOG_LEVEL %q; %w", lvl, err)
	}
	return lvl, nil
}

// reads log level from env or returns a default if not specified.
func getLogLevelOrdefault() (zerolog.Level, error) {
	lvl, err := getLogLevel()
	if err != nil {
		return lvl, err
	}
	if lvl == zerolog.NoLevel {
		lvl = zerolog.WarnLevel
	}
	return lvl, nil
}
