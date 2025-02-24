package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

const (
	// Backend
	defaultDisableClientTLS = false
	defaultBtcBackend       = btctypes.Btcd
	defaultNetParams        = btctypes.Simnet
	// RPC endpoint
	defaultRPCBtcNodeHost      = "127.0.01:18556"
	defaultBtcNodeRPCUser      = "rpcuser"
	defaultBtcNodeRPCPass      = "rpcpass"
	defaultBtcNodeEstimateMode = "CONSERVATIVE"
	// ZMQ endpoints
	defaultZmqSeqEndpoint = "tcp://127.0.0.1:29000"
)

var (
	defaultBtcCAFile = filepath.Join(btcutil.AppDataDir("btcd", false), "rpc.cert")
)

// BTCConfig defines configuration for the Bitcoin client
type BTCConfig struct {
	CAFile           string                    `mapstructure:"ca-file"`
	Endpoint         string                    `mapstructure:"endpoint"`
	Username         string                    `mapstructure:"username"`
	Password         string                    `mapstructure:"password"`
	NetParams        string                    `mapstructure:"net-params"`
	BtcBackend       btctypes.SupportedBackend `mapstructure:"btc-backend"`
	ZmqSeqEndpoint   string                    `mapstructure:"zmq-seq-endpoint"`
	DisableClientTLS bool                      `mapstructure:"no-client-tls"`
}

func (cfg *BTCConfig) validateBasicConfig() error {
	if _, ok := btctypes.GetValidNetParams()[cfg.NetParams]; !ok {
		return fmt.Errorf("invalid net params in config file: %s", cfg.NetParams)
	}

	if _, ok := btctypes.GetValidBackends()[cfg.BtcBackend]; !ok {
		return fmt.Errorf("invalid btc backend value in config file: %s", cfg.BtcBackend)
	}

	return nil
}

func (cfg *BTCConfig) validateBitcoindConfig() error {
	if cfg.BtcBackend != btctypes.Bitcoind {
		return nil
	}

	if cfg.ZmqSeqEndpoint == "" {
		return fmt.Errorf(
			"zmq seq endpoint cannot be empt in config file: %s",
			cfg.ZmqSeqEndpoint,
		)
	}

	return nil
}

// Validate does validation checks on bitcoin node configuration values
func (cfg *BTCConfig) Validate() error {
	if err := cfg.validateBasicConfig(); err != nil {
		return err
	}

	err := cfg.validateBitcoindConfig()
	return err
}

// DefaultBTCConfig returns the default values for
func DefaultBTCConfig() BTCConfig {
	return BTCConfig{
		DisableClientTLS: defaultDisableClientTLS,
		CAFile:           defaultBtcCAFile,
		Endpoint:         defaultRPCBtcNodeHost,
		BtcBackend:       defaultBtcBackend,
		NetParams:        defaultNetParams.String(),
		Username:         defaultBtcNodeRPCUser,
		Password:         defaultBtcNodeRPCPass,
		ZmqSeqEndpoint:   defaultZmqSeqEndpoint,
	}
}

// ReadCertFile reads and returns the content of bitcoin RPC's certificate file
func (cfg *BTCConfig) ReadCertFile() []byte {
	return cfg.readCertificateFile(cfg.CAFile)
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
