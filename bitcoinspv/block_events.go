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

	for { // TODO: do we need the for loop when we alrady have the select?
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
				// TODO: think about a better name for it, since here we dont do the full bootstarp (only confirmations depth here)
				// Also add more information about why here we skip the block subscirpitons and why we dont skip it in inital bootstrap.
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

	ib, _, err := r.getAndValidateBlock(blockEvent)
	if err != nil {
		return err
	}

	r.btcCache.Add(ib)

	return r.processBlock(ib)
}

// TODO: probably we should rename it because we only return error on the empty cache, if the block height is wrong we only log a warning.
func (r *Relayer) validateBlockHeight(blockEvent *btctypes.BlockEvent) error {
	//TODO: the name is confusing, we call it lastCachedBlock but we retrive the first block from hash.
	latestCachedBlock := r.btcCache.First()
	if latestCachedBlock == nil {
		err := fmt.Errorf("cache is empty, restart bootstrap process")
		return err
	}
	if blockEvent.Height < latestCachedBlock.BlockHeight {
		r.logger.Warn().Msgf(
			"Connecting block (height: %d, hash: %s) too early, skipping",
			blockEvent.Height,
			blockEvent.BlockHeader.BlockHash().String(),
		)
	}

	return nil
}

// TODO: maybe we should rename the blockEvent to conntectedBlock?
// TODO: lets refactor it and merge this with validate blockHeight
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
	// TODO: the comment should be: If the recived block does not extend the chain in the cache -> restart bootstrap
	parentHash := msgBlock.Header.PrevBlock
	tipCacheBlock := r.btcCache.Tip()
	if parentHash != tipCacheBlock.BlockHash() {
		return nil, nil, fmt.Errorf(
			// TODO: lets return in the error the block hash instead, more info. Height is not enough.
			"cache tip height: %d is outdated for connecting block %d, bootstrap process needs restart",
			tipCacheBlock.BlockHeight, indexedBlock.BlockHeight,
		)
	}

	return indexedBlock, msgBlock, nil
}

// TODO: consider adding small comments on top of the crucial functions to explain what they do.
// TODO: we return an error here but we dont return an error.
func (r *Relayer) processBlock(indexedBlock *types.IndexedBlock) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.Config.ProcessBlockTimeout)
	defer cancel()

	if indexedBlock == nil {
		r.logger.Debug().Msg("No new headers to submit")
		return nil
	}

	headersToProcess := []*types.IndexedBlock{indexedBlock}

	// TODO: maybe we should rename the processBlock to a different name?
	//  also we have the same name processHeaders in multiple places in the code
	if _, err := r.ProcessHeaders(ctx, headersToProcess); err != nil {
		r.logger.Warn().Msgf("Error submitting header: %v", err)
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
	tipCacheBlock := r.btcCache.Tip() //TODO: consider renaming the tip to a different name maybe: last, latest?
	if tipCacheBlock == nil {
		// TODO: wrong error message, no tip block found in cache? WE should reutn the error from
		// Tip(), here just check the error != nil and then propagate it.
		return fmt.Errorf("no blocks found in cache, bootstrap process must be restarted")
	}

	if blockEvent.BlockHeader.BlockHash() != tipCacheBlock.BlockHash() {
		return fmt.Errorf(
			"cache out of sync during block disconnection, bootstrap process needs to be restarted",
		)
	}

	return nil
}
