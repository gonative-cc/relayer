package cli

import (
	"fmt"

	"github.com/gonative-cc/relayer/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func loadConfig(configFile string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.native-relayer")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	log.Info().Msg("Loaded Configuration:")
	log.Info().Interface("config", config).Msg("")

	return &config, nil
}

func prepareEnv(cmd *cobra.Command) (*Config, error) {
	flags := cmd.Root().PersistentFlags()
	lvl, err := flags.GetString("log-level")
	if err != nil {
		return nil, fmt.Errorf("error getting log level: %w", err)
	}
	logLvl, err := zerolog.ParseLevel(lvl)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	env.InitLogger(logLvl)
	configFile, err := flags.GetString("config")
	if err != nil {
		return nil, fmt.Errorf("error getting config file path: %w", err)
	}
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return config, nil
}
