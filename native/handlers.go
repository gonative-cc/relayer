//revive:disable

package native

import (
	"context"
)

// HandleNewBlock handles the receive of new block from the chain.
func (i *Indexer) HandleNewBlock(ctx context.Context, b Block) error {
	// todo set current block

	// and continues to handle a block normally.
	// return i.HandleBlock(ctx, blk)
	return nil
}

// HandleBlock handles the receive of a block from the chain.
func (i *Indexer) HandleBlock(ctx context.Context, b Block) error {
	return nil
}

// HandleTx handles the receive of new Tx from the chain.
func (i *Indexer) HandleTx(ctx context.Context, blockHeight, blockTimeUnix int, tx Tx) error {
	return nil
}
