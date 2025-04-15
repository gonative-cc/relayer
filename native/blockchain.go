package native

import (
	"context"
)

// Tx is a universal transaction type. Note: should be updated
type Tx []byte

// Block is a universal blockchain block type
type Block interface {
	Transactions() []Tx
}

// Blockchain is the expected blockchain interface the indexer needs to store data in the database.
type Blockchain interface {
	Close(ctx context.Context) error
	ChainID() string
	ChainHeader() (chainID string, latestBlock uint64, err error)
	SubscribeNewBlock(ctx context.Context) (<-chan Block, error)
	Block(ctx context.Context, height int64) (blk Block, err error)
}
