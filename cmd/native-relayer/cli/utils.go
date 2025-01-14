package cli

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func loadConfig(configFile string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %w", err)
		} else {
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
