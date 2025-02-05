package btcwrapper

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"go.uber.org/zap"

	"github.com/gonative-cc/relayer/bitcoinspv"
	relayerconfig "github.com/gonative-cc/relayer/bitcoinspv/config"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
	zmqclient "github.com/gonative-cc/relayer/btcwrapper/zmq"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// NewWithBlockSubscriber creates a new BTC client that subscribes
// to newly connected/disconnected blocks used by spv relayer
func NewWithBlockSubscriber(
	config *relayerconfig.BTCConfig,
	retrySleepDuration,
	maxRetrySleepDuration time.Duration,
	parentLogger *zap.Logger,
) (*Client, error) {
	client, err := initializeClient(config, retrySleepDuration, maxRetrySleepDuration)
	if err != nil {
		return nil, err
	}

	if err := configureClientLogger(client, parentLogger); err != nil {
		return nil, err
	}

	if err := setupBackendConnection(client); err != nil {
		return nil, err
	}

	client.logger.Info("Successfully created the BTC client and connected to the BTC server")

	return client, nil
}

func initializeClient(
	config *relayerconfig.BTCConfig,
	retrySleepDuration time.Duration,
	maxRetrySleepDuration time.Duration,
) (*Client, error) {
	client := &Client{
		blockEventsChannel:    make(chan *relayertypes.BlockEvent, 10000),
		config:                config,
		retrySleepDuration:    retrySleepDuration,
		maxRetrySleepDuration: maxRetrySleepDuration,
	}

	params, err := GetBTCNodeParams(config.NetParams)
	if err != nil {
		return nil, err
	}
	client.chainParams = params

	return client, nil
}

func configureClientLogger(client *Client, parentLogger *zap.Logger) error {
	client.logger = parentLogger.With(zap.String("module", "btcwrapper")).Sugar()
	return nil
}

func setupBackendConnection(client *Client) error {
	switch client.config.BtcBackend {
	case relayertypes.Bitcoind:
		return setupBitcoindConnection(client)
	case relayertypes.Btcd:
		return setupBtcdConnection(client)
	default:
		return fmt.Errorf("unsupported backend type: %v", client.config.BtcBackend)
	}
}

func setupBitcoindConnection(client *Client) error {
	connectionCfg := &rpcclient.ConnConfig{
		Host:         client.config.Endpoint,
		HTTPPostMode: true,
		User:         client.config.Username,
		Pass:         client.config.Password,
		DisableTLS:   client.config.DisableClientTLS,
	}

	rpcClient, err := rpcclient.New(connectionCfg, nil)
	if err != nil {
		return err
	}
	client.Client = rpcClient

	backendVersion := rpcclient.BitcoindPost25
	if backendVersion != rpcclient.BitcoindPre19 && backendVersion != rpcclient.BitcoindPre22 &&
		backendVersion != rpcclient.BitcoindPre24 && backendVersion != rpcclient.BitcoindPre25 &&
		backendVersion != rpcclient.BitcoindPost25 {
		return fmt.Errorf("zmq is only supported by bitcoind, but got %v", backendVersion)
	}

	zeromqClient, err := zmqclient.New(
		client.logger.Desugar(), client.config.ZmqSeqEndpoint, client.blockEventsChannel, rpcClient,
	)
	if err != nil {
		return err
	}
	client.zeromqClient = zeromqClient

	return nil
}

func setupBtcdConnection(client *Client) error {
	notificationHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txs []*btcutil.Tx) {
			client.logger.Debugf(
				"Block %v at height %d has been connected at time %v",
				header.BlockHash(), height, header.Timestamp,
			)
			client.blockEventsChannel <- relayertypes.NewBlockEvent(
				relayertypes.BlockConnected, int64(height), header,
			)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			client.logger.Debugf(
				"Block %v at height %d has been disconnected at time %v",
				header.BlockHash(), height, header.Timestamp,
			)
			client.blockEventsChannel <- relayertypes.NewBlockEvent(
				relayertypes.BlockDisconnected, int64(height), header,
			)
		},
	}

	connectionCfg := &rpcclient.ConnConfig{
		Host:         client.config.Endpoint,
		Endpoint:     "ws",
		User:         client.config.Username,
		Pass:         client.config.Password,
		DisableTLS:   client.config.DisableClientTLS,
		Certificates: client.config.ReadCAFile(),
	}

	rpcClient, err := rpcclient.New(connectionCfg, &notificationHandlers)
	if err != nil {
		return err
	}
	client.Client = rpcClient

	backendVersion, err := rpcClient.BackendVersion()
	if err != nil {
		return fmt.Errorf("failed to get BTC backend: %v", err)
	}
	if backendVersion != rpcclient.BtcdPre2401 && backendVersion != rpcclient.BtcdPost2401 {
		return fmt.Errorf("websocket is only supported by btcd, but got %v", backendVersion)
	}

	return nil
}

func (client *Client) SubscribeNewBlocks() {
	switch client.config.BtcBackend {
	case relayertypes.Btcd:
		if err := bitcoinspv.RetryDo(client.retrySleepDuration, client.maxRetrySleepDuration, func() error {
			err := client.NotifyBlocks()
			if err != nil {
				return err
			}
			client.logger.Info("Successfully subscribed to newly connected/disconnected blocks via WebSocket")
			return nil
		}); err != nil {
			panic(err)
		}
	case relayertypes.Bitcoind:
		err := client.zeromqClient.SubscribeSequence()
		if err != nil {
			panic(err)
		}
	}
}

func (client *Client) BlockEventChannel() <-chan *relayertypes.BlockEvent {
	return client.blockEventsChannel
}
