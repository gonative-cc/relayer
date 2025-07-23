package bitcoinspv

import (
	"sync"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/rs/zerolog"
)

// Relayer manages the Bitcoin SPV relayer functionality
//
//nolint:govet
type Relayer struct {
	// Configuration
	Config *config.RelayerConfig
	logger zerolog.Logger

	// Clients
	btcClient     clients.BTCClient
	lcClient      clients.BitcoinSPV
	indexerClient *clients.IndexerClient

	// Cache and state
	btcCache             *types.BTCCache
	btcConfirmationDepth int64

	// Control
	wg              sync.WaitGroup
	isStarted       bool
	quitChannel     chan struct{}
	quitMu          sync.Mutex
	catchupLoopWait time.Duration
}

// New creates and returns a new relayer object
func New(
	cfg *config.RelayerConfig,
	parentLogger zerolog.Logger,
	btcClient clients.BTCClient,
	lcClient clients.BitcoinSPV,
) (*Relayer, error) {
	logger := parentLogger.With().Str("module", "bitcoinspv").Logger()
	relayer := &Relayer{
		Config:               cfg,
		logger:               logger,
		btcClient:            btcClient,
		lcClient:             lcClient,
		indexerClient:        clients.NewIndexerClient(cfg.IndexerURL, logger),
		btcConfirmationDepth: cfg.BTCConfirmationDepth,
		quitChannel:          make(chan struct{}),
		isStarted:            false,
		catchupLoopWait:      10 * time.Second,
	}

	return relayer, nil
}

// Start initializes and launches the SPV relayer goroutines
// for Bitcoin header verification and relay
func (r *Relayer) Start() {
	r.quitMu.Lock()

	if r.isRunning() {
		r.quitMu.Unlock()
		return
	}

	if r.isShutdown() {
		r.restartAfterShutdown()
	}

	r.isStarted = true
	r.quitMu.Unlock() // Unlock before the init so it doesn't block the Stop()
	r.logger.Debug().Msg("Initializing Relayer...")
	r.initializeRelayer()
	r.logger.Info().Msg("Relayer initialization complete and started (listening for new blocks through ZMQ)")
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
	debug := r.logger.Debug()
	debug.Msg("Running bootstrap...")
	r.multitryBootstrap(false)

	debug.Msg("Bootstrap finished. Launching background goroutines...")
	r.wg.Add(1)
	go r.onBlockEvent()
	debug.Msg("Background goroutines launched.")
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
		r.isStarted = false
		r.logger.Info().Msg("Relayer stop signal sent.")
	}
}

// WaitForShutdown waits for all relayer goroutines to complete before returning
func (r *Relayer) WaitForShutdown() {
	r.wg.Wait()
}
