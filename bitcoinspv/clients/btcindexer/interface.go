package btcindexer

import (
	"context"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// Indexer is the interface for a client that sends blocks to the nBTC indexer.
type Indexer interface {
	SendBlocks(ctx context.Context, blocks []*types.IndexedBlock) error
	GetLatestHeight() (int64, error)
}
