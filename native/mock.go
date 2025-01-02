package native

import (
	"context"

	provtypes "github.com/cometbft/cometbft/light/provider"
	tmtypes "github.com/cometbft/cometbft/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// MockBlockchain is a mock implementation of the Blockchain interface.
type MockBlockchain struct{}

// Close is a mock implementation that does nothing.
func (m MockBlockchain) Close(ctx context.Context) error {
	return nil
}

// ChainID is a mock implementation that returns a dummy chain ID.
func (m MockBlockchain) ChainID() string {
	return "mock-chain-id"
}

// ChainHeader is a mock implementation that returns dummy header information.
func (m MockBlockchain) ChainHeader() (string, uint64, error) {
	return "mock-chain-id", 123, nil
}

// SetChainHeader is a mock implementation that does nothing.
func (m MockBlockchain) SetChainHeader(block *tmtypes.Block) {}

// DecodeTx is a mock implementation that returns a dummy transaction.
func (m MockBlockchain) DecodeTx(tx tmtypes.Tx) (sdktypes.Tx, error) {
	return nil, nil
}

// SubscribeNewBlock is a mock implementation that returns a nil channel and no error.
func (m MockBlockchain) SubscribeNewBlock(ctx context.Context) (<-chan *tmtypes.Block, error) {
	return nil, nil
}

// Block is a mock implementation that returns a dummy block.
func (m MockBlockchain) Block(ctx context.Context, height int64) (*tmtypes.Block, int, error) {
	return &tmtypes.Block{
		Header: tmtypes.Header{
			Height: int64(height),
			// TODO: other header fields
		},
		// TODO: other block fields
	}, 0, nil
}

// CheckTx is a mock implementation that returns no error.
func (m MockBlockchain) CheckTx(ctx context.Context, tx tmtypes.Tx) error {
	return nil
}

// LightProvider is a mock implementation that returns a nil provider.
func (m MockBlockchain) LightProvider() provtypes.Provider {
	return nil
}
