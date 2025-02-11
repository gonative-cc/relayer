// Package zmq reference is taken from https://github.com/joakimofv/go-bitcoindclient which is a
// go wrapper around official zmq package https://github.com/pebbe/zmq4
package zmq

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/btcsuite/btcd/rpcclient"
	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/pebbe/zmq4"
	"go.uber.org/zap"
)

var (
	ErrClientClosed = errors.New("client already closed")
)

// Client manages ZMQ subscriptions and communication with a Bitcoin node.
// It handles ZMQ message routing and provides thread-safe access to subscriptions.
// Must be created with New() and cleaned up with Close().
type Client struct {
	// RPC connection
	rpcClient *rpcclient.Client
	logger    *zap.SugaredLogger

	// Lifecycle management
	closed   int32 // Set atomically
	wg       sync.WaitGroup
	quitChan chan struct{}

	// ZMQ configuration
	zeromqEndpoint     string
	blockEventsChannel chan *relayertypes.BlockEvent

	// ZMQ sockets and subscriptions
	zcontext       *zmq4.Context
	zsubscriber    *zmq4.Socket  // Subscriber socket
	subscriptions  Subscriptions // Subscription management
	zbackendsocket *zmq4.Socket  // Backend socket for internal communication
}

// New creates a new zmq client
func New(
	parentLogger *zap.Logger,
	zeromqEndpoint string,
	blockEventsChannel chan *relayertypes.BlockEvent,
	rpcClient *rpcclient.Client,
) (*Client, error) {
	zmqClient := &Client{
		quitChan:           make(chan struct{}),
		rpcClient:          rpcClient,
		zeromqEndpoint:     zeromqEndpoint,
		logger:             parentLogger.With(zap.String("module", "zmq")).Sugar(),
		blockEventsChannel: blockEventsChannel,
	}

	err := zmqClient.initZMQ()
	if err != nil {
		return nil, err
	}

	zmqClient.wg.Add(1)
	go zmqClient.zmqHandler()

	return zmqClient, nil
}

// initZMQ setups the zmq connections with the bitcoin node
func (c *Client) initZMQ() error {
	var err error

	// Initialize ZMQ context
	if c.zcontext, err = zmq4.NewContext(); err != nil {
		return err
	}

	// Setup subscriber socket
	if c.zsubscriber, err = c.zcontext.NewSocket(zmq4.SUB); err != nil {
		return err
	}
	if err = c.zsubscriber.Connect(c.zeromqEndpoint); err != nil {
		return err
	}

	// Setup back socket
	if c.zbackendsocket, err = c.zcontext.NewSocket(zmq4.PAIR); err != nil {
		return err
	}
	if err = c.zbackendsocket.Bind("inproc://channel"); err != nil {
		return err
	}

	// Setup front socket
	zfront, err := c.zcontext.NewSocket(zmq4.PAIR)
	if err != nil {
		return err
	}
	err = zfront.Connect("inproc://channel")
	if err != nil {
		return err
	}

	c.subscriptions.exitedChannel = make(chan struct{})
	c.subscriptions.zfront = zfront

	return nil
}

// Close closes the zmq connections to the bitcoin node
func (c *Client) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return ErrClientClosed
	}
	if c.zcontext != nil {
		err := c.closeZmqContext()
		if err != nil {
			return err
		}
	}
	close(c.quitChan)
	c.wg.Wait()
	return nil
}

// closeZmqContext closes the zmq context
func (c *Client) closeZmqContext() error {
	c.zcontext.SetRetryAfterEINTR(false)

	c.subscriptions.Lock()
	defer c.subscriptions.Unlock()

	select {
	case <-c.subscriptions.exitedChannel:
	default:
		_, err := c.subscriptions.zfront.SendMessage("term")
		if err != nil {
			return err
		}
	}

	<-c.subscriptions.exitedChannel
	return c.zcontext.Term()
}
