package btcwrapper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	btcwire "github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv"

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

// getChainBlocks returns a chain of indexed blocks from the block at baseHeight to the tipBlock
// note: the caller needs to ensure that tipBlock is on the blockchain
func (client *Client) getChainBlocks(
	baseHeight int64,
	tipBlock *relayertypes.IndexedBlock,
) ([]*relayertypes.IndexedBlock, error) {
	tipHeight := tipBlock.Height
	if err := validateBlockHeights(baseHeight, tipHeight); err != nil {
		return nil, err
	}

	chainBlocks := initializeChainBlocks(baseHeight, tipHeight, tipBlock)

	if tipHeight == baseHeight {
		return chainBlocks, nil
	}

	if err := client.populateChainBlocks(chainBlocks, tipBlock); err != nil {
		return nil, err
	}

	return chainBlocks, nil
}

func validateBlockHeights(baseHeight, tipHeight int64) error {
	if tipHeight < baseHeight {
		return fmt.Errorf(
			"the tip block height %v is less than the base height %v",
			tipHeight, baseHeight,
		)
	}
	return nil
}

func initializeChainBlocks(
	baseHeight, tipHeight int64, tipBlock *relayertypes.IndexedBlock,
) []*relayertypes.IndexedBlock {
	blocks := make([]*relayertypes.IndexedBlock, tipHeight-baseHeight+1)
	blocks[len(blocks)-1] = tipBlock
	return blocks
}

func (client *Client) populateChainBlocks(
	blocks []*relayertypes.IndexedBlock, tipBlock *relayertypes.IndexedBlock,
) error {
	// Start from the second to last block and work backwards
	prevBlockHash := &tipBlock.Header.PrevBlock
	for i := len(blocks) - 2; i >= 0; i-- {
		// Get block info for the previous hash
		indexedBlock, msgBlock, err := client.GetBTCBlockByHash(prevBlockHash)
		if err != nil {
			return fmt.Errorf("failed to get block by hash %x: %w", prevBlockHash, err)
		}

		// Store the indexed block and update prevBlockHash for next iteration
		blocks[i] = indexedBlock
		prevBlockHash = &msgBlock.Header.PrevBlock
	}
	return nil
}

func (client *Client) getBestIndexedBlock() (*relayertypes.IndexedBlock, error) {
	tipHash, err := client.getBestBlockHashRetries()
	if err != nil {
		return nil, fmt.Errorf("failed to get the best block: %w", err)
	}

	tipIndexedBlock, _, err := client.GetBTCBlockByHash(tipHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get the block by hash %x: %w", tipHash, err)
	}

	return tipIndexedBlock, nil
}

// GetBTCTailBlocksByHeight retrieves a sequence of blocks from a given base height up to the current chain tip
func (client *Client) GetBTCTailBlocksByHeight(baseHeight int64) ([]*relayertypes.IndexedBlock, error) {
	// Get the current tip block
	tipBlock, err := client.getBestIndexedBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get tip block: %w", err)
	}

	// Validate base height is not greater than tip
	if baseHeight > tipBlock.Height {
		return nil, fmt.Errorf(
			"base height %d exceeds current tip height %d",
			baseHeight, tipBlock.Height,
		)
	}

	// Get chain of blocks from base to tip
	blocks, err := client.getChainBlocks(baseHeight, tipBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain blocks: %w", err)
	}

	return blocks, nil
}
