package btcwrapper

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/rs/zerolog"

	"github.com/gonative-cc/relayer/bitcoinspv"
	zmqclient "github.com/gonative-cc/relayer/bitcoinspv/clients/btcwrapper/zmq"
	relayerconfig "github.com/gonative-cc/relayer/bitcoinspv/config"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// NewClientWithBlockSubscriber creates a new BTC client that subscribes
// to newly connected/disconnected blocks used by spv relayer
// TODO: consider renaming it to a different name because we only listen (maybe client is not the best name).
func NewClientWithBlockSubscriber(
	config *relayerconfig.BTCConfig,
	retrySleepDuration,
	maxRetrySleepDuration time.Duration,
	parentLogger zerolog.Logger,
) (*Client, error) {
	client, err := initializeClient(config, retrySleepDuration, maxRetrySleepDuration)
	if err != nil {
		return nil, err
	}

	configureClientLogger(client, parentLogger)

	if err := setupBackendConnection(client); err != nil {
		return nil, err
	}

	client.logger.Info().Msg("Successfully created the BTC client and connected to the BTC server")

	return client, nil
}

func initializeClient(
	config *relayerconfig.BTCConfig,
	retrySleepDuration time.Duration,
	maxRetrySleepDuration time.Duration,
) (*Client, error) {
	client := &Client{
		blockEventsChannel:    make(chan *btctypes.BlockEvent, 10000),
		config:                config,
		retrySleepDuration:    retrySleepDuration,
		maxRetrySleepDuration: maxRetrySleepDuration,
	}

	//TODO: This looks like its only for btcd.
	params, err := GetBTCNodeParams(config.NetParams)
	if err != nil {
		return nil, err
	}
	client.chainParams = params

	return client, nil
}

func configureClientLogger(client *Client, parentLogger zerolog.Logger) {
	client.logger = parentLogger.With().Str("module", "btcwrapper").Logger()
}

// TODO: we dont need the default, we already check it in config validate; lets use if;else
func setupBackendConnection(client *Client) error {
	switch client.config.BtcBackend {
	case btctypes.Bitcoind:
		return setupBitcoindConnection(client)
	case btctypes.Btcd:
		return setupBtcdConnection(client)
	default:
		return fmt.Errorf("unsupported backend type: %v", client.config.BtcBackend)
	}
}

// TODO: consider returning the client?
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

	// TODO: what is happening here; Check what backend version we need and implement a correct assertion for it.
	// Find out why do we set the backendVersoin to BitcoindPost25
	backendVersion := rpcclient.BitcoindPost25
	if backendVersion != rpcclient.BitcoindPre19 && backendVersion != rpcclient.BitcoindPre22 &&
		backendVersion != rpcclient.BitcoindPre24 && backendVersion != rpcclient.BitcoindPre25 &&
		backendVersion != rpcclient.BitcoindPost25 {
		return fmt.Errorf("zmq is only supported by bitcoind, but got %v", backendVersion)
	}

	zeromqClient, err := zmqclient.New(
		client.logger, client.config.ZmqSeqEndpoint, client.blockEventsChannel, rpcClient,
	)
	if err != nil {
		return err
	}
	client.zeromqClient = zeromqClient

	return nil
}

func setupBtcdConnection(client *Client) error {
	notificationHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, _ []*btcutil.Tx) {
			client.logger.Debug().Msgf(
				"Block %v at height %d has been connected at time %v",
				header.BlockHash(), height, header.Timestamp,
			)
			client.blockEventsChannel <- btctypes.NewBlockEvent(
				btctypes.BlockConnected, int64(height), header,
			)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			client.logger.Debug().Msgf(
				"Block %v at height %d has been disconnected at time %v",
				header.BlockHash(), height, header.Timestamp,
			)
			client.blockEventsChannel <- btctypes.NewBlockEvent(
				btctypes.BlockDisconnected, int64(height), header,
			)
		},
	}

	connectionCfg := &rpcclient.ConnConfig{
		Host:         client.config.Endpoint,
		Endpoint:     "ws",
		User:         client.config.Username,
		Pass:         client.config.Password,
		DisableTLS:   client.config.DisableClientTLS,
		Certificates: client.config.ReadCertFile(),
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
	// TODO: check the if statement. We should discuss and consider dorpping support for btcd
	if backendVersion != rpcclient.BtcdPre2401 && backendVersion != rpcclient.BtcdPost2401 {
		return fmt.Errorf("websocket is only supported by btcd, but got %v", backendVersion)
	}

	return nil
}

// SubscribeNewBlocks create subscription to new block events
// TODO: panic, lets return error instead, add logger to bitcoind case
func (client *Client) SubscribeNewBlocks() {
	switch client.config.BtcBackend {
	case btctypes.Btcd:
		if err := bitcoinspv.RetryDo(client.logger, client.retrySleepDuration, client.maxRetrySleepDuration, func() error {
			err := client.NotifyBlocks()
			if err != nil {
				return err
			}
			client.logger.Info().Msg("Successfully subscribed to newly connected/disconnected blocks via WebSocket")
			return nil
		}); err != nil {
			panic(err)
		}
	case btctypes.Bitcoind:
		err := client.zeromqClient.SubscribeSequence()
		if err != nil {
			panic(err)
		}
	}
}

// BlockEventChannel returns the channel used for zmq block events
func (client *Client) BlockEventChannel() <-chan *btctypes.BlockEvent {
	return client.blockEventsChannel
}
