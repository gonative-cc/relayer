// Copyright (c) 2022-2022 The Babylon developers
// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcwrapper

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"go.uber.org/zap"

	relayerconfig "github.com/gonative-cc/relayer/bitcoinspv/config"
	realyertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
	zeromq "github.com/gonative-cc/relayer/btcwrapper/zmq"
)

var _ BTCClient = &Client{}

// Client maintains an ongoing connection to a Bitcoin RPC server to access
// information about the current state of the best block chain.
type Client struct {
	*rpcclient.Client
	zeromqClient *zeromq.ZMQClient

	// Chain configuration
	chainParams *chaincfg.Params
	config      *relayerconfig.BTCConfig

	// Logging
	logger *zap.SugaredLogger

	// Retry configuration
	retrySleepDuration    time.Duration
	maxRetrySleepDuration time.Duration

	// Channel for notifying new BTC blocks to relayer
	blockEventsChannel chan *realyertypes.BlockEvent
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
