package bitcoinspv

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// onBlockEvent processes block connection and disconnection events received from the Bitcoin client.
func (r *Relayer) onBlockEvent() {
	defer r.wg.Done()

	for {
		select {
		case blockEvent, openChan := <-r.btcClient.BlockEventChannel():
			if !openChan {
				r.logger.Errorf("BTC Block event channel is now closed")
				return
			}

			if err := r.handleBlockEvent(blockEvent); err != nil {
				r.logger.Warnf(
					"Error in event processing: %v, restarting bootstrap",
					err,
				)
				r.multitryBootstrap(true)
			}
		case <-r.quitChan():
			return
		}
	}
}

// handleBlockEvent processes a block event based on its type
func (r *Relayer) handleBlockEvent(blockEvent *types.BlockEvent) error {
	switch blockEvent.EventType {
	case types.BlockConnected:
		return r.onConnectedBlock(blockEvent)
	case types.BlockDisconnected:
		return r.onDisconnectedBlock(blockEvent)
	default:
		return fmt.Errorf("unknown block event type: %v", blockEvent.EventType)
	}
}

// onConnectedBlock handles connected blocks from the BTC client.
// It is invoked when a new connected block is received from the Bitcoin node.
func (r *Relayer) onConnectedBlock(blockEvent *types.BlockEvent) error {
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

func (r *Relayer) validateBlockHeight(blockEvent *types.BlockEvent) error {
	latestCachedBlock := r.btcCache.First()
	if latestCachedBlock == nil {
		err := fmt.Errorf("cache is empty, restart bootstrap process")
		return err
	}
	if blockEvent.Height < latestCachedBlock.Height {
		r.logger.Debugf(
			"Connecting block (height: %d, hash: %s) too early, skipping",
			blockEvent.Height,
			blockEvent.Header.BlockHash().String(),
		)
		return nil
	}

	return nil
}

func (r *Relayer) validateBlockConsistency(blockEvent *types.BlockEvent) error {
	// verify if block is already in cache and check for consistency
	// NOTE: this scenario can occur when bootstrap process starts after BTC block subscription
	if block := r.btcCache.FindBlock(blockEvent.Height); block != nil {
		if block.BlockHash() == blockEvent.Header.BlockHash() {
			r.logger.Debugf(
				"Connecting block (height: %d, hash: %s) already in cache, skipping",
				block.Height,
				block.BlockHash().String(),
			)
			return nil
		}
		return fmt.Errorf(
			"block mismatch at height %d: connecting block hash %s differs from cached block hash %s",
			blockEvent.Height,
			blockEvent.Header.BlockHash().String(),
			block.BlockHash().String(),
		)
	}

	return nil
}

func (r *Relayer) getAndValidateBlock(
	blockEvent *types.BlockEvent,
) (*types.IndexedBlock, *wire.MsgBlock, error) {
	blockHash := blockEvent.Header.BlockHash()
	indexedBlock, msgBlock, err := r.btcClient.GetBTCBlockByHash(&blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"error retrieving block (height: %d, hash: %v) from BTC client: %w",
			blockEvent.Height, blockHash, err,
		)
	}

	// if parent block != cache tip, cache needs update - restart bootstrap
	parentHash := msgBlock.Header.PrevBlock
	tipCacheBlock := r.btcCache.Tip()
	if parentHash != tipCacheBlock.BlockHash() {
		return nil, nil, fmt.Errorf(
			"cache tip height: %d is outdated for connecting block %d, bootstrap process needs restart",
			tipCacheBlock.Height, indexedBlock.Height,
		)
	}

	return indexedBlock, msgBlock, nil
}

func (r *Relayer) processBlock(indexedBlock *types.IndexedBlock) error {
	if indexedBlock == nil {
		r.logger.Debug("No new headers to submit")
		return nil
	}

	headersToProcess := []*types.IndexedBlock{indexedBlock}

	if _, err := r.ProcessHeaders(headersToProcess); err != nil {
		r.logger.Warnf("Error submitting header: %v", err)
	}

	if _, err := r.ProcessTransactions(headersToProcess); err != nil {
		r.logger.Warnf("Error submitting transactions: %v", err)
	}

	return nil
}

// onDisconnectedBlock manages the removal of blocks
// that have been disconnected from the Bitcoin network.
func (r *Relayer) onDisconnectedBlock(blockEvent *types.BlockEvent) error {
	if err := r.checkDisonnected(blockEvent); err != nil {
		return err
	}

	if err := r.btcCache.RemoveLast(); err != nil {
		r.logger.Warnf(
			"Unable to delete last block from cache: %v, bootstrap process must be restarted",
			err,
		)
		return err
	}

	return nil
}

func (r *Relayer) checkDisonnected(blockEvent *types.BlockEvent) error {
	tipCacheBlock := r.btcCache.Tip()
	if tipCacheBlock == nil {
		return fmt.Errorf("no blocks found in cache, bootstrap process must be restarted")
	}

	if blockEvent.Header.BlockHash() != tipCacheBlock.BlockHash() {
		return fmt.Errorf(
			"cache out of sync during block disconnection, bootstrap process needs to be restarted",
		)
	}

	return nil
}
