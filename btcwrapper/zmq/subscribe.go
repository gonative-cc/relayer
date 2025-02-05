package zmq

import (
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
	zeromq "github.com/pebbe/zmq4"
)

var (
	ErrSubIsDisabled    = errors.New("bitcoin node subscription disabled as zmq-endpoint not set")
	ErrSubHasExited     = errors.New("bitcoin node subscription exited")
	ErrSubAlreadyActive = errors.New("bitcoin node subscription already exists")
)

// SequenceMsg is a subscription event coming from a "sequence" ZMQ message.
type SequenceMsg struct {
	Hash  [32]byte // use encoding/hex.EncodeToString() to get it into the RPC method string format.
	Event relayertypes.EventType
}

type Subscriptions struct {
	sync.RWMutex

	exited      chan struct{}
	zfront      *zeromq.Socket
	latestEvent time.Time
	isActive    bool
}

// SubscribeSequence subscribes to the ZMQ "sequence" messages as SequenceMsg items pushed onto the channel.
// Call cancel to cancel the subscription and let the client release the resources. The channel is closed
// when the subscription is canceled or when the client is closed.
func (c *ZMQClient) SubscribeSequence() error {
	if c.zsubscriber == nil {
		return ErrSubIsDisabled
	}
	c.subscriptions.Lock()
	defer c.subscriptions.Unlock()
	select {
	case <-c.subscriptions.exited:
		return ErrSubHasExited
	default:
	}

	if c.subscriptions.isActive {
		return ErrSubAlreadyActive
	}

	if _, err := c.subscriptions.zfront.SendMessage("subscribe", "sequence"); err != nil {
		return err
	}

	c.subscriptions.isActive = true
	return nil
}

func (c *ZMQClient) zmqHandler() {
	c.cleanup()

	zmqPoller := zeromq.NewPoller()
	zmqPoller.Add(c.zsubscriber, zeromq.POLLIN)
	zmqPoller.Add(c.zbackendsocket, zeromq.POLLIN)
OUTER:
	for {
		// Wait forever until a message can be received or the context was canceled.
		polled, err := zmqPoller.Poll(-1)
		if err != nil {
			break OUTER
		}

		for _, p := range polled {
			switch p.Socket {
			case c.zsubscriber:
				if err := handleSubscriberMessage(c); err != nil {
					break OUTER
				}

			case c.zbackendsocket:
				if err := handleBackendMessage(c); err != nil {
					break OUTER
				}
			}
		}
	}

	c.subscriptions.Lock()
	close(c.subscriptions.exited)
	if err := c.subscriptions.zfront.Close(); err != nil {
		c.logger.Errorf("Error closing zfront: %v", err)
		return
	}
	// Close all subscriber channels.
	if c.subscriptions.isActive {
		err := c.zsubscriber.SetUnsubscribe("sequence")
		if err != nil {
			c.logger.Errorf("Error unsubscribing from sequence: %v", err)
			return
		}
	}

	c.subscriptions.Unlock()
}

// handleSubscriberMessage processes messages from the subscriber socket
func handleSubscriberMessage(c *ZMQClient) error {
	message, err := c.zsubscriber.RecvMessage(0)
	if err != nil {
		return err
	}
	c.subscriptions.latestEvent = time.Now()
	if message[0] == "sequence" {
		var sequenceMessage SequenceMsg
		copy(sequenceMessage.Hash[:], message[1])
		switch message[1][32] {
		case 'C':
			sequenceMessage.Event = relayertypes.BlockConnected
		case 'D':
			sequenceMessage.Event = relayertypes.BlockDisconnected
		default:
			return nil
		}

		c.sendBlockEvent(sequenceMessage.Hash[:], sequenceMessage.Event)
	}
	return nil
}

// handleBackendMessage processes messages from the backend socket
func handleBackendMessage(c *ZMQClient) error {
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

func (c *ZMQClient) cleanup() {
	c.wg.Done()
	if err := c.zsubscriber.Close(); err != nil {
		c.logger.Errorf("Error closing ZMQ socket: %v", err)
	}
	if err := c.zbackendsocket.Close(); err != nil {
		c.logger.Errorf("Error closing ZMQ socket: %v", err)
	}
}

func (c *ZMQClient) sendBlockEvent(hash []byte, event relayertypes.EventType) {
	blockHashStr := hex.EncodeToString(hash)
	blockHash, err := chainhash.NewHashFromStr(blockHashStr)
	if err != nil {
		c.logger.Errorf("Failed to parse block hash %v: %v", blockHashStr, err)
		return
	}

	c.logger.Infof("Received zmq sequence message for block %v", blockHashStr)

	indexedBlock, err := c.getBlockByHash(blockHash)
	if err != nil {
		c.logger.Errorf("Failed to get block %v from BTC client: %v", blockHash, err)
		return
	}

	blockEvent := relayertypes.NewBlockEvent(event, indexedBlock.Height, indexedBlock.Header)
	c.blockEventsChannel <- blockEvent
}

func (c *ZMQClient) getBlockByHash(
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
