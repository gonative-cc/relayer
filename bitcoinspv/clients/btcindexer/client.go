package btcindexer

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/rs/zerolog"
)

// Constants for retry logic
const (
	maxRetries     = 4
	initialBackoff = 500 * time.Millisecond
	maxBackoff     = 8 * time.Second
)

// Payload is structure that the indexer expects.
type Payload struct {
	Height      int64  `json:"height"`
	RawBlockHex string `json:"rawBlockHex"`
}

// Client is a client for communicating with the nBTC indexer worker.
type Client struct {
	url    string
	client *http.Client
	logger zerolog.Logger
}

// NewClient creates a new client for the indexer.
func NewClient(url string, parentLogger zerolog.Logger) *Client {
	return &Client{
		url:    fmt.Sprintf("%s/bitcoin/blocks", url),
		client: &http.Client{Timeout: 10 * time.Second},
		logger: parentLogger.With().Str("module", "btcindexer_client").Logger(),
	}
}

// SendBlocks sends a batch of blocks to the indexer.
func (c *Client) SendBlocks(ctx context.Context, blocks []*types.IndexedBlock) error {
	if c == nil || c.client == nil {
		return errors.New("btcindexer.Client is not initialized")
	}

	if len(blocks) == 0 {
		return nil
	}

	payload := make([]Payload, len(blocks))
	for i, block := range blocks {
		var blockBuffer bytes.Buffer
		if err := block.MsgBlock.Serialize(&blockBuffer); err != nil {
			return fmt.Errorf("failed to serialize block %d: %w", block.BlockHeight, err)
		}
		payload[i] = Payload{
			Height:      block.BlockHeight,
			RawBlockHex: hex.EncodeToString(blockBuffer.Bytes()),
		}
	}

	// TODO: do not use json, just raw payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal indexer payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, "PUT", c.url, bytes.NewReader(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create indexer request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", attempt+1, err)
			c.logger.Warn().Err(lastErr).Msg("Retrying indexer call...")
			c.backoff(ctx, attempt)
			continue
		}

		if resp.StatusCode < 300 {
			c.logger.Info().Int("count", len(blocks)).Int("status_code", resp.StatusCode).Msg("Successfully sent blocks to indexer")
			resp.Body.Close()
			return nil
		}

		// Should not retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("indexer returned a non-retryable error: status %d, body: %s", resp.StatusCode, string(body))
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("attempt %d: indexer returned server error: status %d", attempt+1, resp.StatusCode)
			c.logger.Warn().Err(lastErr).Msg("Retrying...")
			resp.Body.Close()
			c.backoff(ctx, attempt)
			continue
		}
		resp.Body.Close()
	}

	return fmt.Errorf("failed to send blocks to indexer after %d attempts: %w", maxRetries+1, lastErr)
}

func (c *Client) backoff(ctx context.Context, attempt int) {
	if attempt >= maxRetries {
		return
	}
	backoff := time.Duration(1<<attempt) * initialBackoff
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	totalBackoff := backoff + jitter

	c.logger.Info().Dur("wait_duration", totalBackoff).Msgf("Waiting before next attempt..")

	select {
	case <-time.After(totalBackoff):
	case <-ctx.Done():
	}
}
