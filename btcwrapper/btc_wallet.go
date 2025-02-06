package btcwrapper

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"go.uber.org/zap"

	relayerconfig "github.com/gonative-cc/relayer/bitcoinspv/config"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
)

// NewWallet creates a new BTC wallet by establishing a connection to either
// a bitcoind or btcd backend based on configuration
func NewWallet(cfg *relayerconfig.BTCConfig, parentLogger *zap.Logger) (*Client, error) {
	params, err := GetBTCNodeParams(cfg.NetParams)
	if err != nil {
		return nil, err
	}

	wallet := &Client{
		config:      cfg,
		chainParams: params,
		logger:      parentLogger.With(zap.String("module", "btcwrapper_wallet")).Sugar(),
	}

	connCfg := createConnConfig(cfg)

	rpcClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	wallet.logger.Infof("Successfully connected to %s backend", cfg.BtcBackend)
	wallet.Client = rpcClient

	return wallet, nil
}

// createConnConfig creates the appropriate RPC connection config based on the backend type
func createConnConfig(cfg *relayerconfig.BTCConfig) *rpcclient.ConnConfig {
	switch cfg.BtcBackend {
	case relayertypes.Bitcoind:
		return &rpcclient.ConnConfig{
			Host:         cfg.Endpoint + "/wallet/" + cfg.WalletName,
			HTTPPostMode: true,
			User:         cfg.Username,
			Pass:         cfg.Password,
			DisableTLS:   cfg.DisableClientTLS,
		}
	case relayertypes.Btcd:
		return &rpcclient.ConnConfig{
			Host:         cfg.WalletEndpoint,
			Endpoint:     "ws",
			User:         cfg.Username,
			Pass:         cfg.Password,
			DisableTLS:   cfg.DisableClientTLS,
			Certificates: cfg.ReadWalletCAFile(),
		}
	default:
		return &rpcclient.ConnConfig{}
	}
}

// CalculateTransactionFee calculates tx fee based on the given fee rate (BTC/kB) and the tx size
func CalculateTransactionFee(feeRateAmount btcutil.Amount, size uint64) (int64, error) {
	// Convert size to KB
	sizeInKB := float64(size) / 1024

	// Calculate fee by multiplying rate by size
	fee := feeRateAmount.MulF64(sizeInKB)

	return int64(fee), nil
}
