package btcindexer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
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
	logger    zerolog.Logger
	apiClient btcindexer.Client
	network   string

	blocksChan chan []*types.IndexedBlock
	done       chan struct{}
}

// NewClient creates a new client for the indexer.
func NewClient(url string, network string, authToken string, parentLogger zerolog.Logger) *Client {
	c := &Client{
		logger:     parentLogger.With().Str("module", "btcindexer_client").Logger(),
		apiClient:  btcindexer.NewClient(url, authToken),
		network:    network,
		blocksChan: make(chan []*types.IndexedBlock, queueSize),
		done:       make(chan struct{}),
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
		}
	}
}

// SendBlocks enqueues blocks for async sending to the indexer.
// It returns an error only if the queue is full.
func (c *Client) SendBlocks(_ context.Context, blocks []*types.IndexedBlock) error {
	if c == nil {
		return errors.New("btcindexer.Client is not initialized")
	}
	if len(blocks) == 0 {
		return nil
	}

	select {
	case c.blocksChan <- blocks:
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
		shouldRetry, err := c.sendAndHandleResponse(payload)
		if err != nil {
			return err
		}

		if !shouldRetry {
			return nil
		}

		lastErr = fmt.Errorf("attempt %d failed, retrying", attempt+1)
		c.logger.Warn().Err(err).Msg("Retrying indexer call...")
		c.backoff(attempt)
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

	// resp.StatusCode >= 500 {
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

func (c *Client) backoff(attempt int) {
	if attempt >= maxRetries {
		return
	}
	backoff := time.Duration(1<<attempt) * initialBackoff
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond //nolint:gosec
	totalBackoff := backoff + jitter

	c.logger.Info().Dur("wait_duration", totalBackoff).Msgf("Waiting before next attempt..")
	time.Sleep(totalBackoff)
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
func (c *Client) Close() {
	if c == nil {
		return
	}
	close(c.blocksChan)
	<-c.done
}

// Flush waits for all queued blocks to be sent.
func (c *Client) Flush() {
	for len(c.blocksChan) > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}
