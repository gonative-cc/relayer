package bitcoinspv

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// blockEventHandler handles connected and disconnected blocks from the BTC client.
func (r *Relayer) blockEventHandler() {
	defer r.wg.Done()
	quit := r.quitChan()

	for {
		select {
		case event, open := <-r.btcClient.BlockEventChan():
			if !open {
				r.logger.Errorf("block event channel is closed")
				return
			}

			if err := r.handleBlockEvent(event); err != nil {
				r.logger.Warnf(
					"cue to error in event processing: %v, bootstrap process need to be restarted",
					err,
				)
				r.bootstrapWithRetries(true)
			}

		case <-quit:
			return
		}
	}
}

// handleBlockEvent processes a block event based on its type
func (r *Relayer) handleBlockEvent(event *types.BlockEvent) error {
	switch event.EventType {
	case types.BlockConnected:
		return r.handleConnectedBlocks(event)
	case types.BlockDisconnected:
		return r.handleDisconnectedBlocks(event)
	default:
		return fmt.Errorf("unknown block event type: %v", event.EventType)
	}
}

// handleConnectedBlocks handles connected blocks from the BTC client.
func (r *Relayer) handleConnectedBlocks(event *types.BlockEvent) error {
	if err := r.validateBlockHeight(event); err != nil {
		return err
	}

	if err := r.validateBlockConsistency(event); err != nil {
		return err
	}

	ib, _, err := r.getAndValidateBlock(event)
	if err != nil {
		return err
	}

	r.btcCache.Add(ib)

	return r.processBlock(ib)
}

func (r *Relayer) validateBlockHeight(event *types.BlockEvent) error {
	firstCacheBlock := r.btcCache.First()
	if firstCacheBlock == nil {
		return fmt.Errorf("cache is empty, restart bootstrap process")
	}
	if event.Height < firstCacheBlock.Height {
		r.logger.Debugf(
			"the connecting block (height: %d, hash: %s) is too early, skipping the block",
			event.Height,
			event.Header.BlockHash().String(),
		)
		return nil
	}

	return nil
}

func (r *Relayer) validateBlockConsistency(event *types.BlockEvent) error {
	// if the received header is within the cache's region, then this means the events have
	// an overlap with the cache. Then, perform a consistency check. If the block is duplicated,
	// then ignore the block, otherwise there is an inconsistency and redo bootstrap
	// NOTE: this might happen when bootstrapping is triggered after the relayer
	// has subscribed to the BTC blocks
	if b := r.btcCache.FindBlock(event.Height); b != nil {
		if b.BlockHash() == event.Header.BlockHash() {
			r.logger.Debugf(
				"the connecting block (height: %d, hash: %s) is known to cache, skipping the block",
				b.Height,
				b.BlockHash().String(),
			)
			return nil
		}
		return fmt.Errorf(
			"the connecting block (height: %d, hash: %s) is different from the "+
				"header (height: %d, hash: %s) at the same height in cache",
			event.Height,
			event.Header.BlockHash().String(),
			b.Height,
			b.BlockHash().String(),
		)
	}

	return nil
}

func (r *Relayer) getAndValidateBlock(event *types.BlockEvent) (*types.IndexedBlock, *wire.MsgBlock, error) {
	blockHash := event.Header.BlockHash()
	ib, msgBlock, err := r.btcClient.GetBlockByHash(&blockHash)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to get block %v with number %d, from BTC client: %w",
			blockHash, event.Height, err,
		)
	}

	// if the parent of the block is not the tip of the cache, then the cache is not up-to-date,
	// and we might have missed some blocks. In this case, restart the bootstrap process.
	parentHash := msgBlock.Header.PrevBlock
	cacheTip := r.btcCache.Tip() // NOTE: cache is guaranteed to be non-empty at this stage
	if parentHash != cacheTip.BlockHash() {
		return nil, nil, fmt.Errorf(
			"cache (tip %d) is not up-to-date while connecting block %d, restart bootstrap process",
			cacheTip.Height, ib.Height,
		)
	}

	return ib, msgBlock, nil
}

func (r *Relayer) processBlock(ib *types.IndexedBlock) error {
	if ib == nil {
		r.logger.Debug("No new headers to submit to Native light client")
		return nil
	}

	headersToProcess := []*types.IndexedBlock{ib}

	// Process headers
	if _, err := r.ProcessHeaders(headersToProcess); err != nil {
		r.logger.Warnf("Failed to submit header: %v", err)
	}

	// NOTE: not copied
	// Process transactions
	if _, err := r.ProcessTransactions(headersToProcess); err != nil {
		r.logger.Warnf("Failed to submit transactions: %v", err)
	}

	return nil
}

// handleDisconnectedBlocks handles disconnected blocks from the BTC client.
func (r *Relayer) handleDisconnectedBlocks(event *types.BlockEvent) error {
	cacheTip := r.btcCache.Tip()
	if cacheTip == nil {
		return fmt.Errorf("cache is empty, restart bootstrap process")
	}

	if event.Header.BlockHash() != cacheTip.BlockHash() {
		return fmt.Errorf("cache is not up-to-date while disconnecting block, restart bootstrap process")
	}

	if err := r.btcCache.RemoveLast(); err != nil {
		r.logger.Warnf("Failed to remove last block from cache: %v, restart bootstrap process", err)
		return err
	}

	return nil
}
