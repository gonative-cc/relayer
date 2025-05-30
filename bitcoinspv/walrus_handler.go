package bitcoinspv

import (
	"encoding/hex"
	"fmt"

	"github.com/gonative-cc/relayer/bitcoinspv/config"
	walrus "github.com/namihq/walrus-go"
	"github.com/rs/zerolog"
)

// WalrusHandler wraps the Walrus client
type WalrusHandler struct {
	client *walrus.Client
	logger zerolog.Logger
	config *config.RelayerConfig
}

// NewWalrusHandler creates and initializes a new WalrusHandler, nil if not enabled
func NewWalrusHandler(cfg *config.RelayerConfig, parentLogger zerolog.Logger) (*WalrusHandler, error) {
	if !cfg.StoreBlocksInWalrus {
		return nil, nil
	}

	logger := parentLogger.With().Str("module", "bitcoinspv/walrus_handler").Logger()
	logger.Info().Msg("Initializing Walrus client...")

	opts := []walrus.ClientOption{}
	if len(cfg.WalrusPublisherURLs) > 0 {
		opts = append(opts, walrus.WithPublisherURLs(cfg.WalrusPublisherURLs))
	}
	if len(cfg.WalrusAggregatorURLs) > 0 {
		opts = append(opts, walrus.WithAggregatorURLs(cfg.WalrusAggregatorURLs))
	}

	walrusClient := walrus.NewClient(opts...)
	if walrusClient == nil {
		return nil, fmt.Errorf("failed to init Walrus")
	}

	logger.Info().Msg("Walrus client init successful")
	return &WalrusHandler{
		client: walrusClient,
		logger: logger,
		config: cfg,
	}, nil
}

// StoreBlock attempts to store the raw block data in Walrus.
func (wh *WalrusHandler) StoreBlock(
	rawBlockData []byte,
	blockHeight int64,
	blockHashStr string,
) (*string, error) {
	rawBlockHex := hex.EncodeToString(rawBlockData)
	// only for debug (block can be very long)
	wh.logger.Debug().Msgf("Storing block: {\n heigh:%d,\n hash:%s,\n raw_block:%s\n} in Walrus...", 
		blockHeight, blockHashStr, rawBlockHex)

	epochs := wh.config.WalrusStorageEpochs
	storeOpts := &walrus.StoreOptions{Epochs: epochs}

	resp, err := wh.client.Store(rawBlockData, storeOpts)
	if err != nil {
		wh.logger.Error().Err(err).Msgf("Failed to store block %d (%s) in Walrus", blockHeight, blockHashStr)
		return nil, err
	}

	var blobID string
	if resp.NewlyCreated != nil {
		blobID = resp.NewlyCreated.BlobObject.BlobID
		wh.logger.Info().Msgf(
			"Block %d (%s) newly stored in Walrus. Blob ID: %s, Cost: %d",
			blockHeight, blockHashStr, blobID, resp.NewlyCreated.Cost,
		)
	} else if resp.AlreadyCertified != nil {
		blobID = resp.AlreadyCertified.BlobID
		wh.logger.Info().Msgf(
			"Block %d (%s) data already stored in Walrus. Blob ID: %s, End Epoch: %d",
			blockHeight, blockHashStr, blobID, resp.AlreadyCertified.EndEpoch,
		)
	} else {
		return nil, fmt.Errorf("unexpected Walrus store response for block %d (%s)", blockHeight, blockHashStr)
	}
	return &blobID, nil
}
