package native2ika

import (
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
)

// BlockFetcher fetches blocks from the Native chain and saves the txs into database.
type BlockFetcher interface {
	FetchBlock(height int64) (*tmtypes.Block, error)
}

// MockBlockFetcher is a mock implementation of BlockFetcher.
type MockBlockFetcher struct {
	Blocks []*tmtypes.Block
}

// FetchBlock returns a mock block from the Blocks slice.
func (m *MockBlockFetcher) FetchBlock(height int64) (*tmtypes.Block, error) {
	if int(height) < len(m.Blocks) {
		return m.Blocks[height], nil
	}
	return nil, fmt.Errorf("block not found")
}

func getMockBlocks() []*tmtypes.Block {
	//TODO: implement
}
