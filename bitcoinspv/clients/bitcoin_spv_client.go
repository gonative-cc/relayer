package clients

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// BlockInfo represents a simplified Bitcoin block with containing only essential information.
type BlockInfo struct {
	Hash   *chainhash.Hash
	Height int64
}

// BitcoinSPVClient defines the interface for interacting
// with a Bitcoin SPV (Simplified Payment Verification) light client.
// It abstracts the underlying SPV client implementation.
type BitcoinSPVClient interface {

	// InsertHeaders adds new Bitcoin block headers to the light client's chain.
	InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error

	// GetChainTip returns the block hash and height of the best (highest height)
	// block header known to the light client.
	GetHeaderChainTip(ctx context.Context) (*BlockInfo, error)

	// ContainsBlock checks if the light client's chain includes a block with the given hash.
	//
	// Returns:
	//	 - (true, nil) if the block is found.
	//   - (false, nil) if the block is not found.
	//   - (false, error) if there's an error during the check
	ContainsBTCBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error)

	// VerifySPVProof verifies an SPV proof against the light client's stored headers.
	//
	// Returns:
	//	 - (1, nil) TODO: decide on what to return (probably 3 different stages, non-existent, submited, confirmed)
	//   - (2, nil) TODO: ditto
	//   - (3, nil) TODO: ditto
	//   - (-1, error) if there's an error during the check
	VerifySPV(ctx context.Context, spvProof *types.SPVProof) (int, error)

	// Stop gracefully shuts down the SPV light client, releasing any resources.
	Stop()
}
