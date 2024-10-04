package native

import (
	"context"

	provtypes "github.com/cometbft/cometbft/light/provider"
	tmtypes "github.com/cometbft/cometbft/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// Blockchain is the expected blockchain interface the indexer needs to store data in the database.
type Blockchain interface {
	Close(ctx context.Context) error
	ChainID() string
	ChainHeader() (chainID string, height uint64, err error)
	SetChainHeader(*tmtypes.Block)
	DecodeTx(tx tmtypes.Tx) (sdktypes.Tx, error)
	SubscribeNewBlock(ctx context.Context) (<-chan *tmtypes.Block, error)
	Block(ctx context.Context, height int64) (blk *tmtypes.Block, minimumBlkHeight int, err error)
	CheckTx(ctx context.Context, tx tmtypes.Tx) (err error)
	LightProvider() (provtypes.Provider)
}
