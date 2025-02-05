package btcwrapper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv"
	"go.uber.org/zap"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// GetTipBlock retrieves the most recent block in the chain with verbose details
func (c *Client) GetTipBlock() (*btcjson.GetBlockVerboseResult, error) {
	tipBTCBlockHash, err := c.GetBestBlockHash()
	if err != nil {
		return nil, err
	}

	tipBTCBlock, err := c.GetBlockVerbose(tipBTCBlockHash)
	if err != nil {
		return nil, err
	}

	return tipBTCBlock, nil
}

// GetBTCTipBlock provides similar functionality with the btcd.rpcclient.GetBTCTipBlock function
// We implement this, because this function is only provided by btcd.
func (c *Client) GetBTCTipBlock() (*chainhash.Hash, int64, error) {
	btcLatestBlockHash, err := c.getBestBlockHashWithRetry()
	if err != nil {
		return nil, 0, err
	}
	btcLatestBlock, err := c.getBlockVerboseWithRetry(btcLatestBlockHash)
	if err != nil {
		return nil, 0, err
	}
	btcLatestBlockHeight := btcLatestBlock.Height
	return btcLatestBlockHash, btcLatestBlockHeight, nil
}

func (c *Client) GetBTCBlockByHash(blockHash *chainhash.Hash) (*types.IndexedBlock, *wire.MsgBlock, error) {
	blockInfo, err := c.getBlockVerboseWithRetry(blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to get block verbose by hash %s: %w",
			blockHash.String(), err,
		)
	}

	mBlock, err := c.getBlockWithRetry(blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block by hash %s: %w", blockHash.String(), err)
	}

	btcTxs := types.GetWrappedTxs(mBlock)
	return types.NewIndexedBlock(blockInfo.Height, &mBlock.Header, btcTxs), mBlock, nil
}

// GetBTCBlockByHeight returns a block with the given height
func (c *Client) GetBTCBlockByHeight(height int64) (*types.IndexedBlock, *wire.MsgBlock, error) {
	blockHash, err := c.getBlockHashWithRetry(height)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block by height %d: %w", height, err)
	}

	mBlock, err := c.getBlockWithRetry(blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block by hash %s: %w", blockHash.String(), err)
	}

	btcTxs := types.GetWrappedTxs(mBlock)

	return types.NewIndexedBlock(height, &mBlock.Header, btcTxs), mBlock, nil
}

func (c *Client) getBestBlockHashWithRetry() (*chainhash.Hash, error) {
	var (
		blockHash *chainhash.Hash
		err       error
	)

	if err := bitcoinspv.RetryDo(c.retrySleepDuration, c.maxRetrySleepDuration, func() error {
		blockHash, err = c.GetBestBlockHash()
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.logger.Debug("failed to query the best block hash", zap.Error(err))
		return nil, err
	}

	return blockHash, nil
}

func (c *Client) getBlockHashWithRetry(height int64) (*chainhash.Hash, error) {
	var (
		blockHash *chainhash.Hash
		err       error
	)

	if err := bitcoinspv.RetryDo(c.retrySleepDuration, c.maxRetrySleepDuration, func() error {
		blockHash, err = c.GetBlockHash(height)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.logger.Debug(
			"failed to query the block hash",
			zap.Int64("height", height), zap.Error(err),
		)
		return nil, err
	}

	return blockHash, nil
}

func (c *Client) getBlockWithRetry(hash *chainhash.Hash) (*wire.MsgBlock, error) {
	var (
		block *wire.MsgBlock
		err   error
	)

	if err := bitcoinspv.RetryDo(c.retrySleepDuration, c.maxRetrySleepDuration, func() error {
		block, err = c.GetBlock(hash)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.logger.Debug(
			"failed to query the block",
			zap.String("hash", hash.String()), zap.Error(err),
		)
		return nil, err
	}

	return block, nil
}

func (c *Client) getBlockVerboseWithRetry(hash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error) {
	var (
		blockVerbose *btcjson.GetBlockVerboseResult
		err          error
	)

	if err := bitcoinspv.RetryDo(c.retrySleepDuration, c.maxRetrySleepDuration, func() error {
		blockVerbose, err = c.GetBlockVerbose(hash)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.logger.Debug(
			"failed to query the block verbose",
			zap.String("hash", hash.String()), zap.Error(err),
		)
		return nil, err
	}

	return blockVerbose, nil
}

// getChainBlocks returns a chain of indexed blocks from the block at baseHeight to the tipBlock
// note: the caller needs to ensure that tipBlock is on the blockchain
func (c *Client) getChainBlocks(
	baseHeight int64,
	tipBlock *types.IndexedBlock,
) ([]*types.IndexedBlock, error) {
	tipHeight := tipBlock.Height
	if tipHeight < baseHeight {
		return nil, fmt.Errorf(
			"the tip block height %v is less than the base height %v",
			tipHeight, baseHeight,
		)
	}

	// the returned blocks include the block at the base height and the tip block
	chainBlocks := make([]*types.IndexedBlock, tipHeight-baseHeight+1)
	chainBlocks[len(chainBlocks)-1] = tipBlock

	if tipHeight == baseHeight {
		return chainBlocks, nil
	}

	prevHash := &tipBlock.Header.PrevBlock
	// minus 2 is because the tip block is already put in the last position of the slice,
	// and it is ensured that the length of chainBlocks is more than 1
	for i := len(chainBlocks) - 2; i >= 0; i-- {
		ib, mb, err := c.GetBTCBlockByHash(prevHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get block by hash %x: %w", prevHash, err)
		}
		chainBlocks[i] = ib
		prevHash = &mb.Header.PrevBlock
	}

	return chainBlocks, nil
}

func (c *Client) getBestIndexedBlock() (*types.IndexedBlock, error) {
	tipHash, err := c.getBestBlockHashWithRetry()
	if err != nil {
		return nil, fmt.Errorf("failed to get the best block: %w", err)
	}
	tipIb, _, err := c.GetBTCBlockByHash(tipHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get the block by hash %x: %w", tipHash, err)
	}

	return tipIb, nil
}

// GetBTCTailBlocksByHeight returns the chain of blocks from the block at baseHeight to the tip
func (c *Client) GetBTCTailBlocksByHeight(baseHeight int64) ([]*types.IndexedBlock, error) {
	tipIb, err := c.getBestIndexedBlock()
	if err != nil {
		return nil, err
	}

	if baseHeight > tipIb.Height {
		return nil, fmt.Errorf(
			"invalid base height %d, should not be higher than tip block %d",
			baseHeight, tipIb.Height,
		)
	}

	return c.getChainBlocks(baseHeight, tipIb)
}
