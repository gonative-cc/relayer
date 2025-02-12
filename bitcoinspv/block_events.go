package bitcoinspv

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// onBlockEvent handles connected and disconnected blocks from the BTC client.
func (r *Relayer) onBlockEvent() {
	defer r.wg.Done()

	for {
		select {
		case event, open := <-r.btcClient.BlockEventChannel():
			if !open {
				r.logger.Errorf("BTC Block event channel is now closed")
				return
			}

			if err := r.handleBlockEvent(event); err != nil {
				r.logger.Warnf(
					"Error in event processing: %v, bootstrap process need to be restarted",
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
			"the connecting block (height: %d, hash: %s) is too early, skipping the block",
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
				"Connecting block (height: %d, hash: %s) is already in cache, skipping",
				block.Height,
				block.BlockHash().String(),
			)
			return nil
		}
		return fmt.Errorf(
			"the connecting block (height: %d, hash: %s) is different from the "+
				"header (height: %d, hash: %s) at the same height in cache",
			blockEvent.Height,
			blockEvent.Header.BlockHash().String(),
			block.Height,
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
			"failed to get block (hash: %v, height: %d), from BTC client: %w",
			blockHash, blockEvent.Height, err,
		)
	}

	// if parent block != cache tip, cache needs update - restart bootstrap
	parentHash := msgBlock.Header.PrevBlock
	tipCacheBlock := r.btcCache.Tip() // NOTE: cache is guaranteed to be non-empty at this stage
	if parentHash != tipCacheBlock.BlockHash() {
		return nil, nil, fmt.Errorf(
			"cache (tip %d) is not up-to-date while connecting block %d, restart bootstrap process",
			tipCacheBlock.Height, indexedBlock.Height,
		)
	}

	return indexedBlock, msgBlock, nil
}

func (r *Relayer) processBlock(ib *types.IndexedBlock) error {
	if ib == nil {
		r.logger.Debug("No new headers to submit to Native light client")
		return nil
	}

	headersToProcess := []*types.IndexedBlock{ib}

	if _, err := r.ProcessHeaders(headersToProcess); err != nil {
		r.logger.Warnf("Failed to submit header: %v", err)
	}

	if _, err := r.ProcessTransactions(headersToProcess); err != nil {
		r.logger.Warnf("Failed to submit transactions: %v", err)
	}

	return nil
}

// onDisconnectedBlock handles disconnected blocks from the BTC client.
// It is invoked when event for a previously sent connected block
// is to be reversed received from the Bitcoin node.
func (r *Relayer) onDisconnectedBlock(blockEvent *types.BlockEvent) error {
	tipCacheBlock := r.btcCache.Tip()
	if tipCacheBlock == nil {
		return fmt.Errorf("cache is empty, restart bootstrap process")
	}

	if blockEvent.Header.BlockHash() != tipCacheBlock.BlockHash() {
		return fmt.Errorf("cache is not up-to-date while disconnecting block, restart bootstrap process")
	}

	if err := r.btcCache.RemoveLast(); err != nil {
		r.logger.Warnf("Failed to remove last block from cache: %v, restart bootstrap process", err)
		return err
	}

	return nil
}
