package btcwrapper

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"go.uber.org/zap"

	relayerconfig "github.com/gonative-cc/relayer/bitcoinspv/config"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
	zeromq "github.com/gonative-cc/relayer/btcwrapper/zmq"
)

var _ BTCClient = &Client{}

// Client maintains an ongoing connection to a Bitcoin RPC server to access
// information about the current state of the best block chain.
type Client struct {
	*rpcclient.Client
	zeromqClient          *zeromq.Client
	chainParams           *chaincfg.Params
	config                *relayerconfig.BTCConfig
	logger                *zap.SugaredLogger
	blockEventsChannel    chan *btctypes.BlockEvent
	retrySleepDuration    time.Duration
	maxRetrySleepDuration time.Duration
}

// Stop gracefully shuts down the client and closes channels
func (client *Client) Stop() {
	if client != nil {
		client.Shutdown()
		if client.blockEventsChannel != nil {
			close(client.blockEventsChannel)
		}
	}
}
