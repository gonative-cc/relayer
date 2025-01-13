package cli

import "time"

// Configuration struct
type Config struct {
	Native  NativeCfg  `mapstructure:"native"`
	Ika     IkaCfg     `mapstructure:"ika"`
	Btc     BitcoinCfg `mapstructure:"bitcoin"`
	Relayer RelayerCfg `mapstructure:"relayer"`
	DB      DBCfg      `mapstructure:"db"`
}

type NativeCfg struct {
	RPC  string `mapstructure:"native_rpc"`
	GRPC string `mapstructure:"native_grpc"`
}

type IkaCfg struct {
	RPC              string `mapstructure:"ika_rpc"`
	SignerMnemonic   string `mapstructure:"ika_signer_mnemonic"`
	NativeLcPackage  string `mapstructure:"ika_native_lc_package"`
	NativeLcModule   string `mapstructure:"ika_native_lc_module"`
	NativeLcFunction string `mapstructure:"ika_native_lc_function"`
	GasAcc           string `mapstructure:"ika_gas_acc"`
	GasBudget        string `mapstructure:"ika_gas_budget"`
}

type BitcoinCfg struct {
	RPCHost               string `mapstructure:"btc_rpc_host"`
	RPCUser               string `mapstructure:"btc_rpc_user"`
	RPCPass               string `mapstructure:"btc_rpc_pass"`
	ConfirmationThreshold uint8  `mapstructure:"btc_confirmation_threshold"`
	HTTPPostMode          bool   `mapstructure:"http_post_mode"`
	DisableTLS            bool   `mapstructure:"disable_tls"`
}

type RelayerCfg struct {
	ProcessTxsInterval   time.Duration `mapstructure:"process_txs_interval"`
	ConfirmTxsInterval   time.Duration `mapstructure:"confirm_txs_interval"`
	SignReqFetchInterval time.Duration `mapstructure:"sign_req_fetch_interval"`
	SignReqFetchFrom     int           `mapstructure:"sign_req_fetch_from"`
	SignReqFetchLimit    int           `mapstructure:"sign_req_fetch_limit"`
}

type DBCfg struct {
	File string `mapstructure:"db_file"`
}
