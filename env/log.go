package env

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	EnvLogLevel = "LOG_LEVEL"
)

// InitLogger setups zerolog global logger
func InitLogger() error {
	lvl, err := getLogLevel()
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(lvl)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return nil
}

func getLogLevel() (zerolog.Level, error) {
	logLevelStr := os.Getenv(EnvLogLevel)
	if logLevelStr == "" {
		return zerolog.WarnLevel, nil
	}
	lvl, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		return lvl, fmt.Errorf("invalid LOG_LEVEL %q; %w", lvl, err)
	}
	return lvl, nil
}
