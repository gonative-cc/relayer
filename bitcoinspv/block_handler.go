package bitcoinspv

import (
	"fmt"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// blockEventHandler handles connected and disconnected blocks from the BTC client.
func (r *Reporter) blockEventHandler() {
	defer r.wg.Done()
	quit := r.quitChan()

	for {
		select {
		case event, open := <-r.btcClient.BlockEventChan():
			if !open {
				r.logger.Errorf("Block event channel is closed")
				return // channel closed
			}

			var errorRequiringBootstrap error
			if event.EventType == types.BlockConnected {
				errorRequiringBootstrap = r.handleConnectedBlocks(event)
			} else if event.EventType == types.BlockDisconnected {
				errorRequiringBootstrap = r.handleDisconnectedBlocks(event)
			}

			if errorRequiringBootstrap != nil {
				r.logger.Warnf(
					"Due to error in event processing: %v, bootstrap process need to be restarted",
					errorRequiringBootstrap,
				)
				r.bootstrapWithRetries(true)
			}

		case <-quit:
			// We have been asked to stop
			return
		}
	}
}

// handleConnectedBlocks handles connected blocks from the BTC client.
func (r *Reporter) handleConnectedBlocks(event *types.BlockEvent) error {
	// if the header is too early, ignore it
	// NOTE: this might happen when bootstrapping is triggered after the reporter
	// has subscribed to the BTC blocks
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

	// if the received header is within the cache's region, then this means the events have
	// an overlap with the cache. Then, perform a consistency check. If the block is duplicated,
	// then ignore the block, otherwise there is an inconsistency and redo bootstrap
	// NOTE: this might happen when bootstrapping is triggered after the reporter
	// has subscribed to the BTC blocks
	if b := r.btcCache.FindBlock(uint64(event.Height)); b != nil {
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

	// get the block from hash
	blockHash := event.Header.BlockHash()
	ib, mBlock, err := r.btcClient.GetBlockByHash(&blockHash)
	if err != nil {
		return fmt.Errorf(
			"failed to get block %v with number %d ,from BTC client: %w",
			blockHash, event.Height, err,
		)
	}

	// if the parent of the block is not the tip of the cache, then the cache is not up-to-date,
	// and we might have missed some blocks. In this case, restart the bootstrap process.
	parentHash := mBlock.Header.PrevBlock
	cacheTip := r.btcCache.Tip() // NOTE: cache is guaranteed to be non-empty at this stage
	if parentHash != cacheTip.BlockHash() {
		return fmt.Errorf(
			"cache (tip %d) is not up-to-date while connecting block %d, restart bootstrap process",
			cacheTip.Height, ib.Height,
		)
	}

	// otherwise, add the block to the cache
	r.btcCache.Add(ib)

	var headersToProcess []*types.IndexedBlock

	headersToProcess = append(headersToProcess, ib)

	if len(headersToProcess) == 0 {
		r.logger.Debug("No new headers to submit to Native light client")
		return nil
	}

	// extracts and submits headers for each blocks in ibs
	_, err = r.ProcessHeaders(headersToProcess)
	if err != nil {
		r.logger.Warnf("Failed to submit header: %v", err)
	}

	// NOTE: not copied
	// extracts and submits transactions for each blocks in ibs
	_, err = r.ProcessTransactions(headersToProcess)
	if err != nil {
		r.logger.Warnf("Failed to submit transactions: %v", err)
	}

	return nil
}

// handleDisconnectedBlocks handles disconnected blocks from the BTC client.
func (r *Reporter) handleDisconnectedBlocks(event *types.BlockEvent) error {
	// get cache tip
	cacheTip := r.btcCache.Tip()
	if cacheTip == nil {
		return fmt.Errorf("cache is empty, restart bootstrap process")
	}

	// if the block to be disconnected is not the tip of the cache, then the cache is not up-to-date,
	if event.Header.BlockHash() != cacheTip.BlockHash() {
		return fmt.Errorf("cache is not up-to-date while disconnecting block, restart bootstrap process")
	}

	// otherwise, remove the block from the cache
	if err := r.btcCache.RemoveLast(); err != nil {
		r.logger.Warnf("Failed to remove last block from cache: %v, restart bootstrap process", err)
		panic(err)
	}

	return nil
}
