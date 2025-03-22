package btcwrapper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	btcwire "github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv"
	"github.com/rs/zerolog/log"

	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
)

// GetTipBlock retrieves the most recent block in the chain with verbose details
func (client *Client) GetTipBlock() (*btcjson.GetBlockVerboseResult, error) {
	tipBTCBlockHash, err := client.GetBestBlockHash()
	if err != nil {
		return nil, err
	}

	tipBTCBlock, err := client.GetBlockVerbose(tipBTCBlockHash)
	if err != nil {
		return nil, err
	}

	return tipBTCBlock, nil
}

// GetBTCTipBlock returns the latest block hash and height from the Bitcoin network
// This provides similar functionality to btcd.rpcclient.GetBTCTipBlock
func (client *Client) GetBTCTipBlock() (*chainhash.Hash, int64, error) {
	hash, err := client.getBestBlockHashRetries()
	if err != nil {
		return nil, 0, err
	}

	block, err := client.getBlockVerboseRetries(hash)
	if err != nil {
		return nil, 0, err
	}

	return hash, block.Height, nil
}

// GetBTCBlockByHash returns the block of given block hash
func (client *Client) GetBTCBlockByHash(
	blockHash *chainhash.Hash,
) (*relayertypes.IndexedBlock, *btcwire.MsgBlock, error) {
	// Get block info and raw block data in parallel using goroutines
	type blockResult struct {
		info  *btcjson.GetBlockVerboseResult
		block *btcwire.MsgBlock
		err   error
	}

	blockInfoChan := make(chan blockResult)
	blockDataChan := make(chan blockResult)

	go func() {
		info, err := client.getBlockVerboseRetries(blockHash)
		blockInfoChan <- blockResult{info: info, err: err}
	}()

	go func() {
		block, err := client.getBlockRetries(blockHash)
		blockDataChan <- blockResult{block: block, err: err}
	}()

	// Wait for both goroutines to complete
	blockInfoRes := <-blockInfoChan
	if blockInfoRes.err != nil {
		return nil, nil, fmt.Errorf(
			"failed to get block verbose by hash %s: %w",
			blockHash.String(), blockInfoRes.err,
		)
	}

	blockDataRes := <-blockDataChan
	if blockDataRes.err != nil {
		return nil, nil, fmt.Errorf(
			"failed to get block by hash %s: %w",
			blockHash.String(), blockDataRes.err,
		)
	}

	btcTxs := relayertypes.GetWrappedTxs(blockDataRes.block)
	return relayertypes.NewIndexedBlock(
		blockInfoRes.info.Height, &blockDataRes.block.Header, btcTxs,
	), blockDataRes.block, nil
}

// GetBTCBlockByHeight returns a block with the given height
func (client *Client) GetBTCBlockByHeight(
	height int64,
) (*relayertypes.IndexedBlock, *btcwire.MsgBlock, error) {
	// Get block hash for the height
	blockHash, err := client.getBlockHashRetries(height)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block by height %d: %w", height, err)
	}

	// Get the full block data
	indexedBlock, msgBlock, err := client.GetBTCBlockByHash(blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block by hash %s: %w", blockHash.String(), err)
	}

	return indexedBlock, msgBlock, nil
}

func (client *Client) getBestBlockHashRetries() (*chainhash.Hash, error) {
	var blockHash *chainhash.Hash

	if err := bitcoinspv.RetryDo(client.retrySleepDuration, client.maxRetrySleepDuration, func() error {
		var err error
		blockHash, err = client.GetBestBlockHash()
		return err
	}); err != nil {
		return nil, err
	}

	return blockHash, nil
}

func (client *Client) getBlockHashRetries(height int64) (*chainhash.Hash, error) {
	var blockHash *chainhash.Hash

	if err := bitcoinspv.RetryDo(client.retrySleepDuration, client.maxRetrySleepDuration, func() error {
		var err error
		blockHash, err = client.GetBlockHash(height)
		return err
	}); err != nil {
		return nil, err
	}

	return blockHash, nil
}

func (client *Client) getBlockRetries(hash *chainhash.Hash) (*btcwire.MsgBlock, error) {
	var block *btcwire.MsgBlock

	if err := bitcoinspv.RetryDo(client.retrySleepDuration, client.maxRetrySleepDuration, func() error {
		var err error
		block, err = client.GetBlock(hash)
		return err
	}); err != nil {
		return nil, err
	}

	return block, nil
}

func (client *Client) getBlockVerboseRetries(
	hash *chainhash.Hash,
) (*btcjson.GetBlockVerboseResult, error) {
	var blockVerbose *btcjson.GetBlockVerboseResult

	if err := bitcoinspv.RetryDo(client.retrySleepDuration, client.maxRetrySleepDuration, func() error {
		var err error
		blockVerbose, err = client.GetBlockVerbose(hash)
		return err
	}); err != nil {
		return nil, err
	}

	return blockVerbose, nil
}

// GetBTCTailBlocksByHeight retrieves a sequence of blocks or block headers
// from a given base height up to the current chain tip, based on the fullBlocks flag.
func (client *Client) GetBTCTailBlocksByHeight(
	baseHeight int64,
	fullBlocks bool,
) ([]*relayertypes.IndexedBlock, error) {
	// Get the current tip block
	_, tipHeight, err := client.GetBTCTipBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get tip block: %w", err)
	}

	// Validate base height is not greater than tip
	if baseHeight > tipHeight {
		return nil, fmt.Errorf(
			"base height %d exceeds current tip height %d",
			baseHeight, tipHeight,
		)
	}

	totalBlocks := tipHeight - baseHeight + 1
	blocks := make([]*relayertypes.IndexedBlock, 0, totalBlocks)

	for height := baseHeight; height <= tipHeight; height++ {
		var indexedBlock *relayertypes.IndexedBlock
		if fullBlocks {
			block, _, err := client.GetBTCBlockByHeight(height)
			if err != nil {
				return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
			}
			indexedBlock = block
		} else {
			header, err := client.GetBTCBlockHeaderByHeight(height)
			if err != nil {
				return nil, fmt.Errorf("failed to get block header at height %d: %w", height, err)
			}
			indexedBlock = relayertypes.NewIndexedBlock(height, header, []*btcutil.Tx{})
		}

		blocks = append(blocks, indexedBlock)

		// Log progress every 1000 blocks/headers.
		if (height-baseHeight+1)%1000 == 0 || height == tipHeight {
			log.Info().Msgf("Fetched %d/%d blocks/headers (fullBlocks: %t)...", height-baseHeight+1, totalBlocks, fullBlocks)
		}
	}

	log.Info().Msgf("Successfully fetched %d blocks/headers.", totalBlocks)
	return blocks, nil
}

// GetBTCBlockHeaderByHeight retrieves only the block header for a given height.
func (client *Client) GetBTCBlockHeaderByHeight(height int64) (*btcwire.BlockHeader, error) {
	blockHash, err := client.getBlockHashRetries(height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash for height %d: %w", height, err)
	}

	header, err := client.GetBlockHeader(blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block header for hash %s: %w", blockHash.String(), err)
	}

	return header, nil
}
