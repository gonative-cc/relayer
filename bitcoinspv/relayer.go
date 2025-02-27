package bitcoinspv

import (
	"sync"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"go.uber.org/zap"
)

// Relayer manages the Bitcoin SPV relayer functionality
//
//nolint:govet
type Relayer struct {
	// Configuration
	Config *config.RelayerConfig
	logger *zap.SugaredLogger

	// Clients
	btcClient clients.BTCClient
	lcClient  clients.BitcoinSPV

	// Retry settings
	retrySleepDuration    time.Duration
	maxRetrySleepDuration time.Duration

	// Cache and state
	btcCache             *types.BTCCache
	btcConfirmationDepth int64

	// Context timeout
	processBlockTimeout time.Duration

	// Control
	wg          sync.WaitGroup
	isStarted   bool
	quitChannel chan struct{}
	quitMu      sync.Mutex
}

// New creates and returns a new relayer object
func New(
	config *config.RelayerConfig,
	parentLogger *zap.Logger,
	btcClient clients.BTCClient,
	lcClient clients.BitcoinSPV,
	retrySleepDuration,
	maxRetrySleepDuration time.Duration,
	processBlockTimeout time.Duration,
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
		processBlockTimeout:   processBlockTimeout,
		btcClient:             btcClient,
		lcClient:              lcClient,
		btcConfirmationDepth:  defaultConfirmationDepth,
		quitChannel:           make(chan struct{}),
	}

	return relayer, nil
}

// Start initializes and launches the SPV relayer goroutines
// for Bitcoin header verification and relay
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

	r.logger.Infof("Successfully started the spv relayer")
}

// quitChan returns the quit channel in a thread-safe manner.
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
		close(r.quitChannel)
	}
}

// WaitForShutdown waits for all relayer goroutines to complete before returning
func (r *Relayer) WaitForShutdown() {
	r.wg.Wait()
}
