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
	RawBlockHex string `json:"rawBlockHex"`
	Height      int64  `json:"height"`
}

// Client is a client for communicating with the nBTC indexer worker.
type Client struct {
	logger zerolog.Logger
	client *http.Client
	url    string
}

// NewClient creates a new client for the indexer.
func NewClient(url string, parentLogger zerolog.Logger) *Client {
	return &Client{
		url:    fmt.Sprintf("%s/bitcoin/blocks", url),
		client: &http.Client{Timeout: 10 * time.Second},
		logger: parentLogger.With().Str("module", "btcindexer_client").Logger(),
	}
}

// SendBlocks sends a batch of blocks to the indexer with a retry mechanism.
func (c *Client) SendBlocks(ctx context.Context, blocks []*types.IndexedBlock) error {
	if c == nil || c.client == nil {
		return errors.New("btcindexer.Client is not initialized")
	}
	if len(blocks) == 0 {
		return nil
	}

	jsonData, err := c.preparePayload(blocks)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		shouldRetry, err := c.sendAndHandleResponse(ctx, jsonData)
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

func (c *Client) sendAndHandleResponse(ctx context.Context, payload []byte) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url, bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("failed to create indexer request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		c.logger.Info().
			Int("status_code", resp.StatusCode).
			Msgf("Successfully sent %d blocks to indexer", len(payload))
		return false, nil
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("indexer returned a non-retryable error: status %d, body: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode >= 500 {
		return true, fmt.Errorf("indexer returned server error: status %d", resp.StatusCode)
	}

	return true, fmt.Errorf("indexer returned unhandled status: %d", resp.StatusCode)
}

func (c *Client) preparePayload(blocks []*types.IndexedBlock) ([]byte, error) {
	payload := make([]Payload, len(blocks))
	for i, block := range blocks {
		var blockBuffer bytes.Buffer
		if err := block.MsgBlock.Serialize(&blockBuffer); err != nil {
			return nil, fmt.Errorf("failed to serialize block %d: %w", block.BlockHeight, err)
		}
		payload[i] = Payload{
			Height:      block.BlockHeight,
			RawBlockHex: hex.EncodeToString(blockBuffer.Bytes()),
		}
	}
	return json.Marshal(payload)
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
