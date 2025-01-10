package cli

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

func loadConfig() (*Config, error) {
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to the config file")
	flag.Parse()

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

	fmt.Println("Loaded Configuration:")
	fmt.Println(config)

	return &config, nil
}
