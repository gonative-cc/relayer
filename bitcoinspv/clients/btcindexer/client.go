package btcindexer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/gonative-cc/workers/api/btcindexer"
	"github.com/rs/zerolog"
)

// Constants for retry logic
const (
	maxRetries     = 4
	initialBackoff = 500 * time.Millisecond
	maxBackoff     = 8 * time.Second
	queueSize      = 400
)

// Client is a client for communicating with the nBTC indexer worker.
// It wraps the btcindexer API client to add retry logic and async sending.
type Client struct {
	logger      zerolog.Logger
	apiClient   btcindexer.Client
	network     string
	closed      atomicBool
	sendError   atomicError
	wg          sync.WaitGroup
	closeOnce   sync.Once
	retryCtx    context.Context
	retryCancel context.CancelFunc
	blocksChan  chan []*types.IndexedBlock
	done        chan struct{}
}

type atomicBool struct {
	mu sync.RWMutex
	v  bool
}

func (b *atomicBool) get() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.v
}

func (b *atomicBool) set(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.v = v
}

type atomicError struct {
	mu sync.Mutex
	v  error
}

func (e *atomicError) get() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.v
}

func (e *atomicError) set(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.v = err
}

// NewClient creates a new client for the indexer.
func NewClient(url string, network string, authToken string, parentLogger zerolog.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		logger:      parentLogger.With().Str("module", "btcindexer_client").Logger(),
		apiClient:   btcindexer.NewClient(url, authToken),
		network:     network,
		blocksChan:  make(chan []*types.IndexedBlock, queueSize),
		done:        make(chan struct{}),
		retryCtx:    ctx,
		retryCancel: cancel,
	}
	go c.worker()
	return c
}

// worker processes blocks from the queue in a background goroutine.
func (c *Client) worker() {
	defer close(c.done)
	for blocks := range c.blocksChan {
		if err := c.sendBlocksWithRetry(blocks); err != nil {
			c.logger.Error().Err(err).Msg("Failed to send blocks to indexer")
			c.sendError.set(err)
		}
		c.wg.Done()
	}
}

// SendBlocks enqueues blocks for async sending to the indexer.
// It returns an error only if the queue is full or if the client is closed.
// Note: The context parameter is currently ignored for queue operations but is
// checked for cancellation before enqueueing to respect cancellation semantics.
func (c *Client) SendBlocks(ctx context.Context, blocks []*types.IndexedBlock) error {
	if c == nil {
		return errors.New("btcindexer.Client is not initialized")
	}
	if len(blocks) == 0 {
		return nil
	}
	if c.closed.get() {
		return errors.New("btcindexer.Client is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	select {
	case c.blocksChan <- blocks:
		c.wg.Add(1)
		return nil
	default:
		err := errors.New("indexer queue is full, dropping blocks")
		c.logger.Error().
			Err(err).
			Int("dropped_blocks_count", len(blocks)).
			Int("queue_size", queueSize).
			Msg("Indexer queue is full, dropping blocks")
		return err
	}
}

// sendBlocksWithRetry sends a batch of blocks to the indexer with a retry mechanism.
func (c *Client) sendBlocksWithRetry(blocks []*types.IndexedBlock) error {
	payload, err := c.preparePayload(blocks)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		select {
		case <-c.retryCtx.Done():
			return c.retryCtx.Err()
		default:
		}

		shouldRetry, err := c.sendAndHandleResponse(payload)
		if err != nil {
			return err
		}

		if !shouldRetry {
			return nil
		}

		lastErr = fmt.Errorf("attempt %d failed, retrying", attempt+1)
		c.logger.Warn().Err(lastErr).Msg("Retrying indexer call...")
		if !c.backoff(attempt) {
			return errors.New("backoff interrupted by shutdown")
		}
	}

	return fmt.Errorf("failed to send blocks to indexer after %d attempts: %w", maxRetries+1, lastErr)
}

func (c *Client) sendAndHandleResponse(payload btcindexer.PutBlocksReq) (bool, error) {
	resp, err := c.apiClient.PutBlocks(payload)
	if err != nil {
		c.logger.Warn().Err(err).Msg("Indexer call failed with network error, retry.")
		return true, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Info().
			Int("status_code", resp.StatusCode).
			Msgf("Successfully sent %d blocks to indexer", len(payload))
		return false, nil
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("indexer returned a non-retryable error: status %d, body: %s", resp.StatusCode, string(body))
	}

	c.logger.Warn().
		Int("status_code", resp.StatusCode).
		Msg("Indexer returned a server error retry.")
	return true, nil

}

func (c *Client) preparePayload(blocks []*types.IndexedBlock) (btcindexer.PutBlocksReq, error) {
	putBlocksReq := make(btcindexer.PutBlocksReq, len(blocks))
	for i, block := range blocks {
		var blockBuffer bytes.Buffer
		if err := block.MsgBlock.Serialize(&blockBuffer); err != nil {
			return nil, fmt.Errorf("failed to serialize block %d: %w", block.BlockHeight, err)
		}
		putBlocksReq[i] = btcindexer.PutBlock{
			Network: c.network,
			Height:  block.BlockHeight,
			Block:   blockBuffer.Bytes(),
		}
	}
	return putBlocksReq, nil
}

// backoff sleeps for an exponential backoff duration with jitter.
// Returns false if the context was cancelled (e.g., during shutdown), true otherwise.
func (c *Client) backoff(attempt int) bool {
	if attempt >= maxRetries {
		return true
	}
	backoff := time.Duration(1<<attempt) * initialBackoff
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond //nolint:gosec
	totalBackoff := backoff + jitter

	c.logger.Info().Dur("wait_duration", totalBackoff).Msg("Waiting before next attempt...")

	select {
	case <-c.retryCtx.Done():
		return false
	case <-time.After(totalBackoff):
		return true
	}
}

// GetLatestHeight returns the latest block height known to the indexer
func (c *Client) GetLatestHeight() (int64, error) {
	height, err := c.apiClient.GetLatestHeight(c.network)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get latest height from indexer")
		return 0, err
	}

	return height, nil
}

// Close stops the background worker and waits for it to finish.
// It is safe to call multiple times.
func (c *Client) Close() {
	if c == nil {
		return
	}
	c.closeOnce.Do(func() {
		c.closed.set(true)
		c.retryCancel()
		close(c.blocksChan)
		<-c.done
	})
}

// Flush waits for all queued blocks to be sent.
// Returns an error if the client is nil, closed, or if any blocks failed to send.
func (c *Client) Flush() error {
	if c == nil {
		return errors.New("btcindexer.Client is not initialized")
	}
	if c.closed.get() {
		return errors.New("btcindexer.Client is closed")
	}

	c.wg.Wait()
	return c.sendError.get()
}
