package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// IndexedBlock represents a BTC block with additional metadata including block height
// and transaction details needed for Merkle proof generation
type IndexedBlock struct {
	RawMsgBlock *wire.MsgBlock
	BlockHeight int64
}

// NewIndexedBlock creates a new IndexedBlock instance with the given block height,
// header and transactions
func NewIndexedBlock(
	blockHeight int64,
	rawMsgBlock *wire.MsgBlock,
) *IndexedBlock {
	return &IndexedBlock{
		BlockHeight: blockHeight,
		RawMsgBlock: rawMsgBlock,
	}
}

// BlockHash returns the hash of this block's header
func (indexedBlock *IndexedBlock) BlockHash() chainhash.Hash {
	return indexedBlock.RawMsgBlock.BlockHash()
}
