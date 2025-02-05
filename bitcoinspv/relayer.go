package bitcoinspv

import (
	"sync"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/gonative-cc/relayer/lcclient"
	"go.uber.org/zap"
)

// Relayer manages the Bitcoin SPV relayer functionality
type Relayer struct {
	// Configuration
	Config *config.RelayerConfig
	logger *zap.SugaredLogger

	// Clients
	btcClient    clients.BTCClient
	nativeClient *lcclient.Client

	// Retry settings
	retrySleepDuration    time.Duration
	maxRetrySleepDuration time.Duration

	// Cache and state
	btcCache             *types.BTCCache
	btcConfirmationDepth int64
	metrics              *RelayerMetrics

	// Control
	wg          sync.WaitGroup
	isStarted   bool
	quitChannel chan struct{}
	quitMu      sync.Mutex
}

func New(
	config *config.RelayerConfig,
	parentLogger *zap.Logger,
	btcClient clients.BTCClient,
	nativeClient *lcclient.Client,
	retrySleepDuration,
	maxRetrySleepDuration time.Duration,
	metrics *RelayerMetrics,
) (*Relayer, error) {
	logger := parentLogger.With(zap.String("module", "bitcoinspv")).Sugar()

	// to configure how many blocks needs to be pushed on top
	// to assume it is confirmed (no reorg)
	const defaultConfirmationDepth = int64(1)
	logger.Infof("BTCCheckpoint parameters: k = %d", defaultConfirmationDepth)

	relayer := &Relayer{
		Config:                config,
		logger:                logger,
		retrySleepDuration:    retrySleepDuration,
		maxRetrySleepDuration: maxRetrySleepDuration,
		btcClient:             btcClient,
		nativeClient:          nativeClient,
		btcConfirmationDepth:  defaultConfirmationDepth,
		metrics:               metrics,
		quitChannel:           make(chan struct{}),
	}

	return relayer, nil
}

// Start starts the goroutines necessary to manage a spv relayer.
func (r *Relayer) Start() {
	r.quitMu.Lock()
	defer r.quitMu.Unlock()

	if r.isRunning() {
		return
	}

	if r.isShutdown() {
		r.restartAfterShutdown()
	}

	r.isStarted = true
	r.initializeRelayer()
}

func (r *Relayer) isRunning() bool {
	return r.isStarted
}

func (r *Relayer) isShutdown() bool {
	select {
	case <-r.quitChannel:
		return true
	default:
		return false
	}
}

func (r *Relayer) restartAfterShutdown() {
	r.WaitForShutdown()
	r.quitChannel = make(chan struct{})
}

func (r *Relayer) initializeRelayer() {
	r.multitryBootstrap(false)

	r.wg.Add(1)
	go r.onBlockEvent()

	// start record time-related metrics
	r.metrics.RecordMetrics()

	r.logger.Infof("Successfully started the spv relayer")
}

// quitChan atomically reads the quit channel.
func (r *Relayer) quitChan() <-chan struct{} {
	r.quitMu.Lock()
	defer r.quitMu.Unlock()
	return r.quitChannel
}

// Stop signals all spv relayer goroutines to shutdown.
// if already stopped, this is a no-op.
func (r *Relayer) Stop() {
	r.quitMu.Lock()
	defer r.quitMu.Unlock()

	select {
	case <-r.quitChannel:
		// already stopped
		return
	default:
		// closing the `quit` channel will trigger all select case `<-quit`,
		// and thus making all handler routines to break the for loop.
		close(r.quitChannel)
	}
}

// WaitForShutdown blocks until all spv relayer goroutines have finished executing.
func (r *Relayer) WaitForShutdown() {
	r.wg.Wait()
}
