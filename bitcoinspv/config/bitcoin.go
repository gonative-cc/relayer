package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/lightningnetwork/lnd/lnwallet/chainfee"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// BTCConfig defines configuration for the Bitcoin client
type BTCConfig struct {
	// TLS configuration
	DisableClientTLS bool   `mapstructure:"no-client-tls"`
	CAFile           string `mapstructure:"ca-file"`

	// Node configuration
	Endpoint          string                    `mapstructure:"endpoint"`
	Username          string                    `mapstructure:"username"`
	Password          string                    `mapstructure:"password"`
	ReconnectAttempts int                       `mapstructure:"reconnect-attempts"`
	NetParams         string                    `mapstructure:"net-params"`
	BtcBackend        types.SupportedBtcBackend `mapstructure:"btc-backend"`

	// Wallet configuration
	WalletEndpoint string `mapstructure:"wallet-endpoint"`
	WalletPassword string `mapstructure:"wallet-password"`
	WalletName     string `mapstructure:"wallet-name"`
	WalletCAFile   string `mapstructure:"wallet-ca-file"`

	// Fee configuration
	TxFeeMin       chainfee.SatPerKVByte `mapstructure:"tx-fee-min"`
	TxFeeMax       chainfee.SatPerKVByte `mapstructure:"tx-fee-max"`
	DefaultFee     chainfee.SatPerKVByte `mapstructure:"default-fee"`
	EstimateMode   string                `mapstructure:"estimate-mode"`
	TargetBlockNum int64                 `mapstructure:"target-block-num"`

	// ZMQ configuration
	ZmqSeqEndpoint   string `mapstructure:"zmq-seq-endpoint"`
	ZmqBlockEndpoint string `mapstructure:"zmq-block-endpoint"`
	ZmqTxEndpoint    string `mapstructure:"zmq-tx-endpoint"`
}

func (cfg *BTCConfig) validateBasicConfig() error {
	if cfg.ReconnectAttempts < 0 {
		return errors.New("reconnect-attempts must be non-negative")
	}

	if _, ok := types.GetValidNetParams()[cfg.NetParams]; !ok {
		return errors.New("invalid net params")
	}

	if _, ok := types.GetValidBtcBackends()[cfg.BtcBackend]; !ok {
		return errors.New("invalid btc backend")
	}

	return nil
}

func (cfg *BTCConfig) validateBitcoindConfig() error {
	if cfg.BtcBackend != types.Bitcoind {
		return nil
	}

	if cfg.ZmqBlockEndpoint == "" {
		return errors.New("zmq block endpoint cannot be empty")
	}

	if cfg.ZmqTxEndpoint == "" {
		return errors.New("zmq tx endpoint cannot be empty")
	}

	if cfg.ZmqSeqEndpoint == "" {
		return errors.New("zmq seq endpoint cannot be empty")
	}

	if cfg.EstimateMode != "ECONOMICAL" && cfg.EstimateMode != "CONSERVATIVE" {
		return errors.New("estimate-mode must be either ECONOMICAL or CONSERVATIVE when the backend is bitcoind")
	}

	return nil
}

func (cfg *BTCConfig) validateFeeConfig() error {
	if cfg.TargetBlockNum <= 0 {
		return errors.New("target-block-num should be positive")
	}

	if cfg.TxFeeMax <= 0 {
		return errors.New("tx-fee-max must be positive")
	}

	if cfg.TxFeeMin <= 0 {
		return errors.New("tx-fee-min must be positive")
	}

	if cfg.TxFeeMin > cfg.TxFeeMax {
		return errors.New("tx-fee-min is larger than tx-fee-max")
	}

	if cfg.DefaultFee <= 0 {
		return errors.New("default-fee must be positive")
	}

	if cfg.DefaultFee < cfg.TxFeeMin || cfg.DefaultFee > cfg.TxFeeMax {
		return fmt.Errorf("default-fee should be in the range of [%v, %v]", cfg.TxFeeMin, cfg.TxFeeMax)
	}

	return nil
}

func (cfg *BTCConfig) Validate() error {
	if err := cfg.validateBasicConfig(); err != nil {
		return err
	}

	if err := cfg.validateBitcoindConfig(); err != nil {
		return err
	}

	if err := cfg.validateFeeConfig(); err != nil {
		return err
	}

	return nil
}

const (
	DefaultTxPollingJitter     = 0.5
	DefaultRPCBtcNodeHost      = "127.0.01:18556"
	DefaultBtcNodeRPCUser      = "rpcuser"
	DefaultBtcNodeRPCPass      = "rpcpass"
	DefaultBtcNodeEstimateMode = "CONSERVATIVE"
	DefaultBtcblockCacheSize   = 20 * 1024 * 1024 // 20 MB
	DefaultZmqSeqEndpoint      = "tcp://127.0.0.1:29000"
	DefaultZmqBlockEndpoint    = "tcp://127.0.0.1:29001"
	DefaultZmqTxEndpoint       = "tcp://127.0.0.1:29002"
)

func DefaultBTCConfig() BTCConfig {
	return BTCConfig{
		DisableClientTLS:  false,
		CAFile:            defaultBtcCAFile,
		Endpoint:          DefaultRPCBtcNodeHost,
		WalletEndpoint:    "localhost:18554",
		WalletPassword:    "walletpass",
		WalletName:        "default",
		WalletCAFile:      defaultBtcWalletCAFile,
		BtcBackend:        types.Btcd,
		TxFeeMax:          chainfee.SatPerKVByte(20 * 1000), // 20,000sat/kvb = 20sat/vbyte
		TxFeeMin:          chainfee.SatPerKVByte(1 * 1000),  // 1,000sat/kvb = 1sat/vbyte
		DefaultFee:        chainfee.SatPerKVByte(1 * 1000),  // 1,000sat/kvb = 1sat/vbyte
		EstimateMode:      DefaultBtcNodeEstimateMode,
		TargetBlockNum:    1,
		NetParams:         types.BtcSimnet.String(),
		Username:          DefaultBtcNodeRPCUser,
		Password:          DefaultBtcNodeRPCPass,
		ReconnectAttempts: 3,
		ZmqSeqEndpoint:    DefaultZmqSeqEndpoint,
		ZmqBlockEndpoint:  DefaultZmqBlockEndpoint,
		ZmqTxEndpoint:     DefaultZmqTxEndpoint,
	}
}

func (cfg *BTCConfig) ReadCAFile() []byte {
	if cfg.DisableClientTLS {
		return nil
	}

	certs, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil
	}

	return certs
}

func (cfg *BTCConfig) ReadWalletCAFile() []byte {
	if cfg.DisableClientTLS {
		return nil
	}

	certs, err := os.ReadFile(cfg.WalletCAFile)
	if err != nil {
		return nil
	}
	return certs
}
