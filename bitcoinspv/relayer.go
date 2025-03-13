package bitcoinspv

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"go.uber.org/zap"
)

const lastRelayedHeightFile = "last_relayed_height"

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
	dataDir              string

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
	dataDir string,
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
		dataDir:               dataDir,
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

// loadLastRelayedHeight loads the last relayed height from persistent storage.
func (r *Relayer) loadLastRelayedHeight() (int64, error) {
	fullPath := filepath.Join(r.dataDir, lastRelayedHeightFile)
	data, err := os.ReadFile(fullPath) // Use os.ReadFile
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return a default value (e.g., -1 or 0).
			return -1, nil // -1 indicates no previous height
		}
		return 0, fmt.Errorf("failed to read last relayed height: %w", err)
	}

	height, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse last relayed height: %w", err)
	}

	return height, nil
}

// saveLastRelayedHeight saves the last relayed height to persistent storage.
func (r *Relayer) saveLastRelayedHeight(height int64) error {
	fullPath := filepath.Join(r.dataDir, lastRelayedHeightFile)
	data := []byte(strconv.FormatInt(height, 10))
	err := os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write last relayed height: %w", err)
	}
	return nil
}

// getHeaderHeight retrieves the block height for a given block header.  It uses the BTCClient.
func (r *Relayer) getHeaderHeight(header *wire.BlockHeader) (int64, error) {
	blockHash := header.BlockHash()
	blockInfo, _, err := r.btcClient.GetBTCBlockByHash(&blockHash)
	if err != nil {
		return 0, fmt.Errorf("failed to get block info for hash %s: %w", blockHash.String(), err)
	}
	return blockInfo.BlockHeight, nil
}
