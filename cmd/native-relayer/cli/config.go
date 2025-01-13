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
	HTTPPostMode             bool          `mapstructure:"http_post_mode"`
	DisableTLS               bool          `mapstructure:"disable_tls"`
	ProcessTxsInterval       time.Duration `mapstructure:"process_txs_interval"`
	ConfirmTxsInterval       time.Duration `mapstructure:"confirm_txs_interval"`
	SignReqFetchInterval     time.Duration `mapstructure:"sign_req_fetch_interval"`
	SignReqFetchFrom         int           `mapstructure:"sign_req_fetch_from"`
	SignReqFetchLimit        int           `mapstructure:"sign_req_fetch_limit"`
	DBFile                   string        `mapstructure:"db_file"`
}
