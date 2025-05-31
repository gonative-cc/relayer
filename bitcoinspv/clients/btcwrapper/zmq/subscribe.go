package zmq

import (
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
	zeromq "github.com/pebbe/zmq4"
)

var (
	errSubscriberIsDisabled    = errors.New("bitcoin node subscription disabled as zmq-endpoint not set")
	errSubscriberHasExited     = errors.New("bitcoin node subscription exited")
	errSubscriberAlreadyActive = errors.New("bitcoin node subscription already exists")
)

// SequenceMessage denotes the message struct received from zmq
type SequenceMessage struct {
	Hash  [32]byte // use encoding/hex.EncodeToString() to get it into the RPC method string format.
	Event btctypes.EventType
}

// Subscriptions keeps track of the zmq connection state
//
//nolint:govet
type Subscriptions struct {
	sync.RWMutex
	latestEvent   time.Time
	exitedChannel chan struct{}
	zfront        *zeromq.Socket
	isActive      bool
}

// SubscribeSequence subscribes to ZMQ "sequence" messages. Call cancel to unsubscribe.
func (c *Client) SubscribeSequence() error {
	if c.zsubscriber == nil {
		return errSubscriberIsDisabled
	}
	c.subscriptions.Lock()
	defer c.subscriptions.Unlock()
	select {
	case <-c.subscriptions.exitedChannel:
		return errSubscriberHasExited
	default:
	}

	if c.subscriptions.isActive {
		return errSubscriberAlreadyActive
	}

	if _, err := c.subscriptions.zfront.SendMessage("subscribe", "sequence"); err != nil {
		return err
	}

	c.subscriptions.isActive = true
	return nil
}

func (c *Client) zmqHandlePolled(polled []zeromq.Polled) error {
	var err error
	for _, p := range polled {
		switch p.Socket {
		case c.zsubscriber:
			err = handleSubscriberMessage(c)
		case c.zbackendsocket:
			err = handleBackendMessage(c)
		}

		if err != nil {
			return err
		}
	}
	return err
}

func (c *Client) zmqPoll() {
	defer c.cleanup()

	zmqPoller := zeromq.NewPoller()
	zmqPoller.Add(c.zsubscriber, zeromq.POLLIN)
	zmqPoller.Add(c.zbackendsocket, zeromq.POLLIN)

	for {
		// Wait forever until a message can be received or the context was canceled.
		polled, err := zmqPoller.Poll(-1)
		if err != nil {
			c.logger.Err(err).Msg("polling zmq data")
			break
		}
		if err = c.zmqHandlePolled(polled); err != nil {
			c.logger.Err(err).Msg("can't handle zmq polled data")
			break
		}
	}

	c.subscriptions.Lock()
	close(c.subscriptions.exitedChannel)
	if err := c.subscriptions.zfront.Close(); err != nil {
		c.logger.Err(err).Msg("Error closing zfront")
		return
	}
	// Close all subscriber channels.
	if c.subscriptions.isActive {
		err := c.zsubscriber.SetUnsubscribe("sequence")
		if err != nil {
			c.logger.Err(err).Msgf("Error unsubscribing from sequence")
			return
		}
	}

	c.subscriptions.Unlock()
}

func handleSubscriberMessage(c *Client) error {
	message, err := c.zsubscriber.RecvMessage(0)
	if err != nil {
		return err
	}
	c.subscriptions.latestEvent = time.Now()
	if message[0] == "sequence" {
		var sequenceMessage SequenceMessage
		copy(sequenceMessage.Hash[:], message[1])
		switch message[1][32] {
		case 'C':
			sequenceMessage.Event = btctypes.BlockConnected
		case 'D':
			sequenceMessage.Event = btctypes.BlockDisconnected
		default:
			return nil
		}

		c.sendBlockEventToChannel(sequenceMessage.Hash[:], sequenceMessage.Event)
	}
	return nil
}

func handleBackendMessage(c *Client) error {
	message, err := c.zbackendsocket.RecvMessage(0)
	if err != nil {
		return err
	}
	switch message[0] {
	case "subscribe":
		if err := c.zsubscriber.SetSubscribe(message[1]); err != nil {
			return err
		}
	case "term":
		return errors.New("termination requested")
	}
	return nil
}

func (c *Client) cleanup() {
	c.wg.Done()
	if err := c.zsubscriber.Close(); err != nil {
		c.logger.Err(err).Msg("Error closing ZMQ socket")
	}
	if err := c.zbackendsocket.Close(); err != nil {
		c.logger.Err(err).Msg("Error closing ZMQ socket")
	}
}

func (c *Client) sendBlockEventToChannel(hashBytes []byte, event btctypes.EventType) {
	blockHashString := hex.EncodeToString(hashBytes)
	blockHash, err := chainhash.NewHashFromStr(blockHashString)
	if err != nil {
		c.logger.Err(err).Msg("Failed to parse block hash " + blockHashString)
		return
	}

	c.logger.Info().Msg("Received zmq sequence message for block " + blockHashString)

	indexedBlock, err := c.getBlockByHash(blockHash)
	if err != nil {
		c.logger.Err(err).Msgf("Failed to get block %v from BTC client", blockHash)
		return
	}

	blockEvent := btctypes.NewBlockEvent(event, indexedBlock.BlockHeight, indexedBlock.BlockHeader)
	c.blockEventsChannel <- blockEvent
}

func (c *Client) getBlockByHash(
	blockHash *chainhash.Hash,
) (*relayertypes.IndexedBlock, error) {
	blockVerbose, err := c.rpcClient.GetBlockVerbose(blockHash)
	if err != nil {
		return nil, err
	}

	block, err := c.rpcClient.GetBlock(blockHash)
	if err != nil {
		return nil, err
	}

	btcTxs := relayertypes.GetWrappedTxs(block)
	indexedBlock := relayertypes.NewIndexedBlock(blockVerbose.Height, &block.Header, btcTxs)

	return indexedBlock, nil
}
