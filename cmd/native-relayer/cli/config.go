package cli

import "time"

// Config aggregates configurations for different components within the native-relayer.
type Config struct {
	Ika     IkaCfg     `mapstructure:"ika"`
	Native  NativeCfg  `mapstructure:"native"`
	DB      DBCfg      `mapstructure:"db"`
	Btc     BitcoinCfg `mapstructure:"bitcoin"`
	Relayer RelayerCfg `mapstructure:"relayer"`
}

// NativeCfg holds configuration for interacting with Native.
type NativeCfg struct {
	RPC  string `mapstructure:"rpc"`
	GRPC string `mapstructure:"grpc"`
}

// IkaCfg holds configuration specific to Ika.
type IkaCfg struct {
	RPC              string `mapstructure:"rpc"`
	SignerMnemonic   string `mapstructure:"signer_mnemonic"`
	NativeLcPackage  string `mapstructure:"native_lc_package"`
	NativeLcModule   string `mapstructure:"native_lc_module"`
	NativeLcFunction string `mapstructure:"native_lc_function"`
	GasAcc           string `mapstructure:"gas_acc"`
	GasBudget        string `mapstructure:"gas_budget"`
}

// BitcoinCfg holds configuration for interacting with Bitcoin.
type BitcoinCfg struct {
	RPCHost               string `mapstructure:"rpc_host"`
	RPCUser               string `mapstructure:"rpc_user"`
	RPCPass               string `mapstructure:"rpc_pass"`
	Network               string `mapstructure:"network"`
	ConfirmationThreshold uint8  `mapstructure:"confirmation_threshold"`
	HTTPPostMode          bool   `mapstructure:"http_post_mode"`
	DisableTLS            bool   `mapstructure:"disable_tls"`
}

// RelayerCfg holds configuration specific to the relayer's operation.
type RelayerCfg struct {
	ProcessTxsInterval   time.Duration `mapstructure:"process_txs_interval"`
	ConfirmTxsInterval   time.Duration `mapstructure:"confirm_txs_interval"`
	SignReqFetchInterval time.Duration `mapstructure:"sign_req_fetch_interval"`
	SignReqFetchFrom     int           `mapstructure:"sign_req_fetch_from"`
	SignReqFetchLimit    int           `mapstructure:"sign_req_fetch_limit"`
}

// DBCfg holds configuration for the database.
type DBCfg struct {
	File string `mapstructure:"file"`
}
