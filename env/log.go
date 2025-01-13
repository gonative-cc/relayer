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
func InitLogger(lvl *zerolog.Level) error {
	var err error
	var loglevel zerolog.Level
	if lvl == nil {
		loglevel, err = getLogLevel()
		if err != nil {
			return err
		}
	}

	zerolog.SetGlobalLevel(loglevel)
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
