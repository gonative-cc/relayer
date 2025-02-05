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
	zmqEndpoint    string
	blockEventChan chan *relayertypes.BlockEvent

	// ZMQ sockets and subscriptions
	zcontext       *zmq4.Context
	zsubscriber    *zmq4.Socket  // Subscriber socket
	subs           subscriptions // Subscription management
	zbackendsocket *zmq4.Socket  // Backend socket for internal communication
}

func New(
	parentLogger *zap.Logger,
	zmqEndpoint string,
	blockEventChan chan *relayertypes.BlockEvent,
	rpcClient *rpcclient.Client,
) (*Client, error) {
	c := &Client{
		quitChan:       make(chan struct{}),
		rpcClient:      rpcClient,
		zmqEndpoint:    zmqEndpoint,
		logger:         parentLogger.With(zap.String("module", "zmq")).Sugar(),
		blockEventChan: blockEventChan,
	}

	if err := c.initZMQ(); err != nil {
		return nil, err
	}

	c.wg.Add(1)
	go c.zmqHandler()

	return c, nil
}

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
	if err = c.zsubscriber.Connect(c.zmqEndpoint); err != nil {
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
	if err = zfront.Connect("inproc://channel"); err != nil {
		return err
	}

	c.subs.exited = make(chan struct{})
	c.subs.zfront = zfront

	return nil
}

func (c *Client) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return ErrClientClosed
	}
	if c.zcontext != nil {
		if err := c.closeZmqContext(); err != nil {
			return err
		}
	}
	close(c.quitChan)
	c.wg.Wait()
	return nil
}

func (c *Client) closeZmqContext() error {
	c.zcontext.SetRetryAfterEINTR(false)

	c.subs.Lock()
	defer c.subs.Unlock()

	select {
	case <-c.subs.exited:
	default:
		if _, err := c.subs.zfront.SendMessage("term"); err != nil {
			return err
		}
	}

	<-c.subs.exited
	return c.zcontext.Term()
}
