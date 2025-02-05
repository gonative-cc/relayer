package bitcoinspv

import (
	"sync"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"go.uber.org/zap"
)

// NOTE: modified
// Relayer manages the Bitcoin SPV relayer functionality
type Relayer struct {
	// Configuration
	Cfg    *config.RelayerConfig
	logger *zap.SugaredLogger

	// Clients
	btcClient    clients.BTCClient
	nativeClient clients.NativeClient

	// Retry settings
	retrySleepTime    time.Duration
	maxRetrySleepTime time.Duration

	// Cache and state
	btcCache             *types.BTCCache
	btcConfirmationDepth int64
	metrics              *RelayerMetrics

	// Control
	wg      sync.WaitGroup
	started bool
	quit    chan struct{}
	quitMu  sync.Mutex
}

func New(
	cfg *config.RelayerConfig,
	parentLogger *zap.Logger,
	btcClient clients.BTCClient,
	nativeClient clients.NativeClient,
	retrySleepTime,
	maxRetrySleepTime time.Duration,
	metrics *RelayerMetrics,
) (*Relayer, error) {
	logger := parentLogger.With(zap.String("module", "bitcoinspv")).Sugar()

	const defaultConfirmationDepth = int64(1)
	logger.Infof("BTCCheckpoint parameters: k = %d", defaultConfirmationDepth)

	relayer := &Relayer{
		Cfg:                  cfg,
		logger:               logger,
		retrySleepTime:       retrySleepTime,
		maxRetrySleepTime:    maxRetrySleepTime,
		btcClient:            btcClient,
		nativeClient:         nativeClient,
		btcConfirmationDepth: defaultConfirmationDepth,
		metrics:              metrics,
		quit:                 make(chan struct{}),
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

	r.started = true
	r.initializeRelayer()
}

func (r *Relayer) isRunning() bool {
	return r.started
}

func (r *Relayer) isShutdown() bool {
	select {
	case <-r.quit:
		return true
	default:
		return false
	}
}

func (r *Relayer) restartAfterShutdown() {
	r.WaitForShutdown()
	r.quit = make(chan struct{})
}

func (r *Relayer) initializeRelayer() {
	r.multitryBootstrap(false)

	r.wg.Add(1)
	go r.onBlockEvent()

	// start record time-related metrics
	r.metrics.RecordMetrics()

	r.logger.Infof("successfully started the spv relayer")
}

// quitChan atomically reads the quit channel.
func (r *Relayer) quitChan() <-chan struct{} {
	r.quitMu.Lock()
	c := r.quit
	r.quitMu.Unlock()
	return c
}

// Stop signals all spv relayer goroutines to shutdown.
// if already stopped, this is a no-op.
func (r *Relayer) Stop() {
	r.quitMu.Lock()
	defer r.quitMu.Unlock()

	select {
	case <-r.quit:
		// already stopped
		return
	default:
		// closing the `quit` channel will trigger all select case `<-quit`,
		// and thus making all handler routines to break the for loop.
		close(r.quit)
	}
}

// ShuttingDown returns whether the spv relayer is currently in the process of shutting down or not.
func (r *Relayer) ShuttingDown() bool {
	select {
	case <-r.quitChan():
		return true
	default:
		return false
	}
}

// WaitForShutdown blocks until all spv relayer goroutines have finished executing.
func (r *Relayer) WaitForShutdown() {
	// TODO: let Native client WaitForShutDown
	r.wg.Wait()
}
