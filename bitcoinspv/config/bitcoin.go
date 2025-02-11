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
		return errors.New("config file value reconnect-attempts must be non-negative")
	}

	if _, ok := types.GetValidNetParams()[cfg.NetParams]; !ok {
		return errors.New("invalid net params in config file")
	}

	if _, ok := types.GetValidBtcBackends()[cfg.BtcBackend]; !ok {
		return errors.New("invalid btc backend value in config file")
	}

	return nil
}

func (cfg *BTCConfig) validateBitcoindConfig() error {
	if cfg.BtcBackend != types.Bitcoind {
		return nil
	}

	if cfg.ZmqBlockEndpoint == "" {
		return errors.New("zmq block endpoint cannot be empty in config file")
	}

	if cfg.ZmqTxEndpoint == "" {
		return errors.New("zmq tx endpoint cannot be empty in config file")
	}

	if cfg.ZmqSeqEndpoint == "" {
		return errors.New("zmq seq endpoint cannot be empt in config file")
	}

	if cfg.EstimateMode != "ECONOMICAL" && cfg.EstimateMode != "CONSERVATIVE" {
		return errors.New("estimate-mode must be in (ECONOMICAL, CONSERVATIVE) when the backend is bitcoind")
	}

	return nil
}

func (cfg *BTCConfig) validateFeeConfig() error {
	if cfg.TargetBlockNum <= 0 {
		return errors.New("target-block-num should be positive in config file")
	}

	if cfg.TxFeeMax <= 0 {
		return errors.New("tx-fee-max must be positive in config file")
	}

	if cfg.TxFeeMin <= 0 {
		return errors.New("tx-fee-min must be positive in config file")
	}

	if cfg.TxFeeMin > cfg.TxFeeMax {
		return errors.New("tx-fee-min is larger than tx-fee-max in config file")
	}

	if cfg.DefaultFee <= 0 {
		return errors.New("default-fee must be positive in config file")
	}

	if cfg.DefaultFee < cfg.TxFeeMin || cfg.DefaultFee > cfg.TxFeeMax {
		return fmt.Errorf("default-fee should be in the range of [%v, %v]", cfg.TxFeeMin, cfg.TxFeeMax)
	}

	return nil
}

// Validate does validation checks on bitcoin node configuration values
func (cfg *BTCConfig) Validate() error {
	if err := cfg.validateBasicConfig(); err != nil {
		return err
	}

	if err := cfg.validateBitcoindConfig(); err != nil {
		return err
	}

	err := cfg.validateFeeConfig()
	return err
}

const (
	// RPC endpoint
	defaultRPCBtcNodeHost      = "127.0.01:18556"
	defaultBtcNodeRPCUser      = "rpcuser"
	defaultBtcNodeRPCPass      = "rpcpass"
	defaultBtcNodeEstimateMode = "CONSERVATIVE"
	// ZMQ endpoints
	defaultZmqSeqEndpoint   = "tcp://127.0.0.1:29000"
	defaultZmqBlockEndpoint = "tcp://127.0.0.1:29001"
	defaultZmqTxEndpoint    = "tcp://127.0.0.1:29002"
)

// DefaultBTCConfig returns the default values for
func DefaultBTCConfig() BTCConfig {
	return BTCConfig{
		DisableClientTLS:  false,
		CAFile:            defaultBtcCAFile,
		Endpoint:          defaultRPCBtcNodeHost,
		WalletEndpoint:    "localhost:18554",
		WalletPassword:    "walletpass",
		WalletName:        "default",
		WalletCAFile:      defaultBtcWalletCAFile,
		BtcBackend:        types.Btcd,
		TxFeeMax:          chainfee.SatPerKVByte(20 * 1000), // 20,000sat/kvb = 20sat/vbyte
		TxFeeMin:          chainfee.SatPerKVByte(1 * 1000),  // 1,000sat/kvb = 1sat/vbyte
		DefaultFee:        chainfee.SatPerKVByte(1 * 1000),  // 1,000sat/kvb = 1sat/vbyte
		EstimateMode:      defaultBtcNodeEstimateMode,
		TargetBlockNum:    1,
		NetParams:         types.BtcSimnet.String(),
		Username:          defaultBtcNodeRPCUser,
		Password:          defaultBtcNodeRPCPass,
		ReconnectAttempts: 3,
		ZmqSeqEndpoint:    defaultZmqSeqEndpoint,
		ZmqBlockEndpoint:  defaultZmqBlockEndpoint,
		ZmqTxEndpoint:     defaultZmqTxEndpoint,
	}
}

// ReadCAFile reads and returns the content of bitcoin RPC's certificate file
func (cfg *BTCConfig) ReadCAFile() []byte {
	return cfg.readCertificateFile(cfg.CAFile)
}

// ReadWalletCAFile reads and returns the content of bitcoin wallet RPC's certificate file
func (cfg *BTCConfig) ReadWalletCAFile() []byte {
	return cfg.readCertificateFile(cfg.WalletCAFile)
}

func (cfg *BTCConfig) readCertificateFile(filePath string) []byte {
	if cfg.DisableClientTLS {
		return nil
	}

	certs, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return certs
}
