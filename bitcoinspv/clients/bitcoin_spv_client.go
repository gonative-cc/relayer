package clients

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// BlockInfo represents a simplified Bitcoin block with containing only essential information.
type BlockInfo struct {
	Hash   *chainhash.Hash
	Height int64
}

// BitcoinSPV defines the interface for interacting
// with a Bitcoin SPV (Simplified Payment Verification) light client.
// It abstracts the underlying SPV client implementation.
type BitcoinSPV interface {

	// InsertHeaders adds new Bitcoin block headers to the light client's chain.
	InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error

	// GetLatestBlockInfo returns the block hash and height of the best (highest height)
	// block header known to the light client.
	GetLatestBlockInfo(ctx context.Context) (*BlockInfo, error)

	// ContainsBlock checks if the light client's chain includes a block with the given hash.
	//
	// Returns:
	//	 - (true, nil) if the block is found.
	//   - (false, nil) if the block is not found.
	//   - (false, error) if there's an error during the check
	ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error)

	// Stop gracefully shuts down the SPV light client, releasing any resources.
	Stop()
}
