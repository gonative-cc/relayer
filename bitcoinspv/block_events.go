package bitcoinspv

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"

	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

// onBlockEvent processes block connection and disconnection events received from the Bitcoin client.
func (r *Relayer) onBlockEvent() {
	defer r.wg.Done()

	for {
		select {
		case blockEvent, openChan := <-r.btcClient.BlockEventChannel():
			if !openChan {
				r.logger.Error().Msg("BTC Block event channel is now closed")
				return
			}

			if err := r.handleBlockEvent(blockEvent); err != nil {
				r.logger.Warn().Msgf(
					"Error in event processing: %v, restarting bootstrap",
					err,
				)
				// We call the bootstrap here, every time there is an error when processing evetns.
				// This usually happens when there is a reorg and the new blocks that are received
				// through ZMQ do not extend the chain known to the lgiht client.
				// Then we need to re-send all the blocks starting from the latest common ancestor.
				r.multitryBootstrap(true)
			}
		case <-r.quitChan():
			return
		}
	}
}

// handleBlockEvent processes a block event based on its type
func (r *Relayer) handleBlockEvent(blockEvent *btctypes.BlockEvent) error {
	switch blockEvent.Type {
	case btctypes.BlockConnected:
		return r.onConnectedBlock(blockEvent)
	case btctypes.BlockDisconnected:
		return r.onDisconnectedBlock(blockEvent)
	default:
		return fmt.Errorf("unknown block event type: %v", blockEvent.Type)
	}
}

// onConnectedBlock handles connected blocks from the BTC client.
// It is invoked when a new connected block is received from the Bitcoin node.
func (r *Relayer) onConnectedBlock(blockEvent *btctypes.BlockEvent) error {
	if err := r.validateBlockHeight(blockEvent); err != nil {
		return err
	}

	if err := r.validateBlockConsistency(blockEvent); err != nil {
		return err
	}

	ib, msgBlock, err := r.getAndValidateBlock(blockEvent)
	if err != nil {
		return err
	}

	// Store full block in Walrus
	if r.Config.StoreBlocksInWalrus && r.walrusHandler != nil {
		r.UploadToWalrus(msgBlock, ib.BlockHeight, ib.BlockHash().String())
	}

	r.btcCache.Add(ib)

	return r.processBlock(ib)
}

func (r *Relayer) validateBlockHeight(blockEvent *btctypes.BlockEvent) error {
	latestCachedBlock := r.btcCache.First()
	if latestCachedBlock == nil {
		err := fmt.Errorf("cache is empty, restart bootstrap process")
		return err
	}
	if blockEvent.Height < latestCachedBlock.BlockHeight {
		r.logger.Debug().Msgf(
			"Connecting block (height: %d, hash: %s) too early, skipping",
			blockEvent.Height,
			blockEvent.BlockHeader.BlockHash().String(),
		)
		return nil
	}

	return nil
}

func (r *Relayer) validateBlockConsistency(blockEvent *btctypes.BlockEvent) error {
	// verify if block is already in cache and check for consistency
	// NOTE: this scenario can occur when bootstrap process starts after BTC block subscription
	if block := r.btcCache.FindBlock(blockEvent.Height); block != nil {
		if block.BlockHash() == blockEvent.BlockHeader.BlockHash() {
			r.logger.Debug().Msgf(
				"Connecting block (height: %d, hash: %s) already in cache, skipping",
				block.BlockHeight,
				block.BlockHash().String(),
			)
			return nil
		}
		return fmt.Errorf(
			"block mismatch at height %d: connecting block hash %s differs from cached block hash %s",
			blockEvent.Height,
			blockEvent.BlockHeader.BlockHash().String(),
			block.BlockHash().String(),
		)
	}

	return nil
}

func (r *Relayer) getAndValidateBlock(
	blockEvent *btctypes.BlockEvent,
) (*types.IndexedBlock, *wire.MsgBlock, error) {
	blockHash := blockEvent.BlockHeader.BlockHash()
	indexedBlock, msgBlock, err := r.btcClient.GetBTCBlockByHash(&blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"error retrieving block (height: %d, hash: %v) from BTC client: %w",
			blockEvent.Height, blockHash, err,
		)
	}

	// if parent block != cache tip, cache needs update - restart bootstrap
	parentHash := msgBlock.Header.PrevBlock
	tipCacheBlock := r.btcCache.Last()
	if parentHash != tipCacheBlock.BlockHash() {
		return nil, nil, fmt.Errorf(
			"cache tip height: %d is outdated for connecting block %d, bootstrap process needs restart",
			tipCacheBlock.BlockHeight, indexedBlock.BlockHeight,
		)
	}

	return indexedBlock, msgBlock, nil
}

func (r *Relayer) processBlock(indexedBlock *types.IndexedBlock) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.Config.ProcessBlockTimeout)
	defer cancel()

	if indexedBlock == nil {
		r.logger.Debug().Msg("No new headers to submit")
		return nil
	}

	headersToProcess := []*types.IndexedBlock{indexedBlock}

	if _, err := r.ProcessHeaders(ctx, headersToProcess); err != nil {
		return err
	}

	return nil
}

// onDisconnectedBlock manages the removal of blocks
// that have been disconnected from the Bitcoin network.
func (r *Relayer) onDisconnectedBlock(blockEvent *btctypes.BlockEvent) error {
	if err := r.checkDisonnected(blockEvent); err != nil {
		return err
	}

	if err := r.btcCache.RemoveLast(); err != nil {
		r.logger.Warn().Msgf(
			"Unable to delete last block from cache: %v, bootstrap process must be restarted",
			err,
		)
		return err
	}

	return nil
}

func (r *Relayer) checkDisonnected(blockEvent *btctypes.BlockEvent) error {
	tipCacheBlock := r.btcCache.Last()
	if tipCacheBlock == nil {
		return fmt.Errorf("no blocks found in cache, bootstrap process must be restarted")
	}

	if blockEvent.BlockHeader.BlockHash() != tipCacheBlock.BlockHash() {
		return fmt.Errorf(
			"cache out of sync during block disconnection, bootstrap process needs to be restarted",
		)
	}

	return nil
}
