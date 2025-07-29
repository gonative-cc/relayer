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
)

// Client is a client for communicating with the nBTC indexer worker.
// It wraps the btcindexer API client to add retry logic.
type Client struct {
	logger    zerolog.Logger
	apiClient btcindexer.Client
}

// NewClient creates a new client for the indexer.
func NewClient(url string, parentLogger zerolog.Logger) *Client {
	return &Client{
		logger:    parentLogger.With().Str("module", "btcindexer_client").Logger(),
		apiClient: btcindexer.NewClient(url),
	}
}

// SendBlocks sends a batch of blocks to the indexer with a retry mechanism.
// TODO: this should not block the main process
// probably we should use CF queues
func (c *Client) SendBlocks(ctx context.Context, blocks []*types.IndexedBlock) error {
	if c == nil {
		return errors.New("btcindexer.Client is not initialized")
	}
	if len(blocks) == 0 {
		return nil
	}

	payload, err := c.preparePayload(blocks)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		shouldRetry, err := c.sendAndHandleResponse(payload)
		if err != nil {
			// Non-retryable
			return err
		}

		if !shouldRetry {
			// Success
			return nil
		}

		lastErr = fmt.Errorf("attempt %d failed, retrying", attempt+1)
		c.logger.Warn().Err(err).Msg("Retrying indexer call...")
		c.backoff(ctx, attempt)
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
			Height: block.BlockHeight,
			Block:  blockBuffer.Bytes(),
		}
	}
	return putBlocksReq, nil
}

func (c *Client) backoff(ctx context.Context, attempt int) {
	if attempt >= maxRetries {
		return
	}
	backoff := time.Duration(1<<attempt) * initialBackoff
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	// NOTE: we dont need secure random generation here, its just retry
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond //nolint:gosec
	totalBackoff := backoff + jitter

	c.logger.Info().Dur("wait_duration", totalBackoff).Msgf("Waiting before next attempt..")

	select {
	case <-time.After(totalBackoff):
	case <-ctx.Done():
	}
}
