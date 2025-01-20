package bitcoinspv

import (
	"sync"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/gonative-cc/relayer/lcclient"
	"go.uber.org/zap"
)

// NOTE: modified
type Relayer struct {
	Cfg    *config.RelayerConfig
	logger *zap.SugaredLogger

	btcClient    BTCClient
	nativeClient *lcclient.Client

	// retry attributes
	retrySleepTime    time.Duration
	maxRetrySleepTime time.Duration

	// Internal states of the relayer
	// CheckpointCache *types.CheckpointCache
	btcCache             *types.BTCCache
	btcConfirmationDepth int64
	metrics              *RelayerMetrics
	wg                   sync.WaitGroup
	started              bool
	quit                 chan struct{}
	quitMu               sync.Mutex
}

func New(
	cfg *config.RelayerConfig,
	parentLogger *zap.Logger,
	btcClient BTCClient,
	nativeClient *lcclient.Client,
	retrySleepTime,
	maxRetrySleepTime time.Duration,
	metrics *RelayerMetrics,
) (*Relayer, error) {
	logger := parentLogger.With(zap.String("module", "bitcoinspv")).Sugar()

	k := int64(1)
	logger.Infof("BTCCheckpoint parameters: k = %d", k)

	return &Relayer{
		Cfg:               cfg,
		logger:            logger,
		retrySleepTime:    retrySleepTime,
		maxRetrySleepTime: maxRetrySleepTime,
		btcClient:         btcClient,
		nativeClient:      nativeClient,
		// CheckpointCache:   ckptCache,
		btcConfirmationDepth: k,
		metrics:              metrics,
		quit:                 make(chan struct{}),
	}, nil
}

// Start starts the goroutines necessary to manage a spv relayer.
func (r *Relayer) Start() {
	r.quitMu.Lock()
	select {
	case <-r.quit:
		// Restart the spv relayer goroutines after shutdown finishes.
		r.WaitForShutdown()
		r.quit = make(chan struct{})
	default:
		// Ignore when the spv relayer is still running.
		if r.started {
			r.quitMu.Unlock()
			return
		}
		r.started = true
	}
	r.quitMu.Unlock()

	r.bootstrapWithRetries(false)

	r.wg.Add(1)
	go r.blockEventHandler()

	// start record time-related metrics
	r.metrics.RecordMetrics()

	r.logger.Infof("Successfully started the spv relayer")
}

// quitChan atomically reads the quit channel.
func (r *Relayer) quitChan() <-chan struct{} {
	r.quitMu.Lock()
	c := r.quit
	r.quitMu.Unlock()
	return c
}

// Stop signals all spv relayer goroutines to shutdown.
func (r *Relayer) Stop() {
	r.quitMu.Lock()
	quit := r.quit
	r.quitMu.Unlock()

	select {
	case <-quit:
	default:
		// closing the `quit` channel will trigger all select case `<-quit`,
		// and thus making all handler routines to break the for loop.
		close(quit)
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
