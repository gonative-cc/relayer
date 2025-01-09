package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Configuration struct to match your YAML structure
type Config struct {
	NativeRPC                string        `mapstructure:"native_rpc"`
	NativeGRPC               string        `mapstructure:"native_grpc"`
	StartBlockHeight         int           `mapstructure:"start_block_height"`
	IkaRPC                   string        `mapstructure:"ika_rpc"`
	IkaSignerMnemonic        string        `mapstructure:"ika_signer_mnemonic"`
	IkaNativeLcPackage       string        `mapstructure:"ika_native_lc_package"`
	IkaNativeLcModule        string        `mapstructure:"ika_native_lc_module"`
	IkaNativeLcFunction      string        `mapstructure:"ika_native_lc_function"`
	IkaGasAcc                string        `mapstructure:"ika_gas_acc"`
	IkaGasBudget             uint64        `mapstructure:"ika_gas_budget"`
	BtcRpcHost               string        `mapstructure:"btc_rpc_host"`
	BtcRpcUser               string        `mapstructure:"btc_rpc_user"`
	BtcRpcPass               string        `mapstructure:"btc_rpc_pass"`
	BtcConfirmationThreshold uint8         `mapstructure:"btc_confirmation_threshold"`
	ProcessTxsInterval       time.Duration `mapstructure:"process_txs_interval"`
	ConfirmTxsInterval       time.Duration `mapstructure:"confirm_txs_interval"`
	DbFile                   string        `mapstructure:"db_file"`
}

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
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
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
