package bitcoinspv

import (
	"context"
	"fmt"

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
	if err := r.checkBlockValidity(blockEvent); err != nil {
		return err
	}

	ib := new(types.IndexedBlock)
	// Store full block in Walrus
	if r.Config.StoreBlocksInWalrus && r.walrusHandler != nil {
		h := blockEvent.BlockHeader.BlockHash()
		ib, err := r.btcClient.GetBTCBlockByHash(&h)
		if err != nil {
			return err
		}
		r.UploadToWalrus(ib.MsgBlock, ib.BlockHeight, ib.BlockHash().String())
	} else {
		ib.BlockHeight = blockEvent.Height
		ib.MsgBlock.Header = *blockEvent.BlockHeader
	}

	err := r.btcCache.Add(ib)
	if err != nil {
		return err
	}

	return r.processBlock(ib)
}

// checkBlockValidity checks the status of a new block
// Steps:
//  1. Checks if cache is empty
//  2. Skips verify if a new block not old enough (new block heigh < first block in cache)
//  3. Checks if appending a new block to cache is possible
//  4. Checks reorg happen, If reorg happened we will rebootstrap relayer
func (r *Relayer) checkBlockValidity(b *btctypes.BlockEvent) error {
	if r.btcCache.IsEmpty() {
		return fmt.Errorf("cache is empty, restart bootstrap process")
	}

	f := r.btcCache.First()
	if f.BlockHeight > b.Height {
		r.logger.Debug().Msgf(
			"Connecting block (height: %d, hash: %s) too early, skipping",
			b.Height,
			b.BlockHeader.BlockHash().String(),
		)
		return nil
	}

	// check if we can append a new block to cache
	l := r.btcCache.Last()
	if l.BlockHeight+1 == b.Height {
		if l.BlockHash() == b.BlockHeader.PrevBlock {
			return nil
		}
		return fmt.Errorf(
			"cache tip height: %d is outdated for connecting block %d, bootstrap process needs restart",
			l.BlockHeight, b.Height,
		)
	}

	reOrg, err := r.isReOrg(b)
	if err != nil {
		return err
	}
	if reOrg {
		return fmt.Errorf("reorg happened at block heigh %d, rebootstrap relayer", b.Height)
	}
	return nil
}

// isReorg checks if the block is a part of new chain after re-org
func (r *Relayer) isReOrg(b *btctypes.BlockEvent) (bool, error) {
	cb, err := r.btcCache.FindBlock(b.Height)
	if err != nil {
		return false, fmt.Errorf("can't find new block in cache %w", err)
	}
	if cb.BlockHash() != b.BlockHeader.BlockHash() {
		return true, nil
	}
	return false, nil
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
