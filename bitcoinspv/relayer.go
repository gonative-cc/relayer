package bitcoinspv

import (
	"sync"

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
	btcClient clients.BTCClient
	lcClient  clients.BitcoinSPV

	// Cache and state
	btcCache             *types.BTCCache
	btcConfirmationDepth int64

	// Control
	wg          sync.WaitGroup
	isStarted   bool
	quitChannel chan struct{}
	quitMu      sync.Mutex
}

// New creates and returns a new relayer object
func New(
	config *config.RelayerConfig,
	parentLogger zerolog.Logger,
	btcClient clients.BTCClient,
	lcClient clients.BitcoinSPV,
) (*Relayer, error) {
	logger := parentLogger.With().Str("module", "bitcoinspv").Logger()
	relayer := &Relayer{
		Config:               config,
		logger:               logger,
		btcClient:            btcClient,
		lcClient:             lcClient,
		btcConfirmationDepth: config.BTCConfirmationDepth,
		quitChannel:          make(chan struct{}),
		isStarted:            false,
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
	r.logger.Debug().Msg("Bootstrap finished.")

	r.logger.Debug().Msg("Launching background goroutines...")
	r.wg.Add(1)
	go r.onBlockEvent()
	r.logger.Debug().Msg("Background goroutines launched.")
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
