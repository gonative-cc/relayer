package cli

import "time"

// Configuration struct
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
	IkaGasBudget             string        `mapstructure:"ika_gas_budget"`
	BtcRPCHost               string        `mapstructure:"btc_rpc_host"`
	BtcRPCUser               string        `mapstructure:"btc_rpc_user"`
	BtcRPCPass               string        `mapstructure:"btc_rpc_pass"`
	BtcConfirmationThreshold uint8         `mapstructure:"btc_confirmation_threshold"`
	ProcessTxsInterval       time.Duration `mapstructure:"process_txs_interval"`
	ConfirmTxsInterval       time.Duration `mapstructure:"confirm_txs_interval"`
	DBFile                   string        `mapstructure:"db_file"`
}
