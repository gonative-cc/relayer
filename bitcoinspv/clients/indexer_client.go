package clients

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/rs/zerolog"
)

// IndexerPayload is structure that the indexer expects.
type IndexerPayload struct {
	Height      int64  `json:"height"`
	RawBlockHex string `json:"rawBlockHex"`
}

// IndexerClient is a client for communicating with the nBTC indexer worker.
type IndexerClient struct {
	url    string
	client *http.Client
	logger zerolog.Logger
}

// NewIndexerClient creates a new client for the indexer.
func NewIndexerClient(url string, parentLogger zerolog.Logger) *IndexerClient {
	return &IndexerClient{
		url:    fmt.Sprintf("%s/bitcoin/blocks", url),
		client: &http.Client{Timeout: 10 * time.Second},
		logger: parentLogger.With().Str("module", "indexer_client").Logger(),
	}
}

// SendBlocks sends a batch of blocks to the indexer.
func (c *IndexerClient) SendBlocks(blocks []*types.IndexedBlock) {
	if len(blocks) == 0 {
		return
	}

	payload := make([]IndexerPayload, len(blocks))
	for i, block := range blocks {
		var blockBuffer bytes.Buffer
		if err := block.MsgBlock.Serialize(&blockBuffer); err != nil {
			c.logger.Error().Err(err).Msgf("Failed to serialize block %d for indexer", block.BlockHeight)
			return
		}
		payload[i] = IndexerPayload{
			Height:      block.BlockHeight,
			RawBlockHex: hex.EncodeToString(blockBuffer.Bytes()),
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to marshal indexer payload to JSON")
		return
	}

	req, err := http.NewRequest("PUT", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create indexer request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to send blocks to indexer")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		c.logger.Error().Int("status_code", resp.StatusCode).Msg("Indexer returned non-success status code")
		return
	}

	c.logger.Info().Int("count", len(blocks)).Msg("Successfully sent blocks to indexer")
}
