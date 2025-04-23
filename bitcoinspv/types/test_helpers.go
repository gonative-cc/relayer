package types

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/wire"
)

// CreateTestIndexedBlocks is a test helper function that generates a slice of mock IndexedBlock objects.
func CreateTestIndexedBlocks(t *testing.T, count int64, startHeight int64) []*IndexedBlock {
	t.Helper()
	blocks := make([]*IndexedBlock, count)
	for i := int64(0); i < count; i++ {
		//nolint: gosec // This is a test function, and overflow is unlikely.
		hdr := wire.BlockHeader{Version: int32(i + 1), Timestamp: time.Now().Add(time.Duration(i) * time.Minute)}
		blocks[i] = &IndexedBlock{
			BlockHeight: startHeight + i,
			BlockHeader: &hdr,
		}
	}
	return blocks
}
