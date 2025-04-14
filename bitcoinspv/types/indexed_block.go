package types

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// IndexedBlock represents a BTC block with additional metadata including block height
// and transaction details needed for Merkle proof generation
type IndexedBlock struct {
	BlockHeader  *wire.BlockHeader
	Transactions []*btcutil.Tx
	BlockHeight  int64
}

// NewIndexedBlock creates a new IndexedBlock instance with the given block height,
// header and transactions
// TODO: think about renaming it (maybe light block?)
func NewIndexedBlock(
	blockHeight int64,
	blockHeader *wire.BlockHeader,
	transactions []*btcutil.Tx,
) *IndexedBlock {
	return &IndexedBlock{
		BlockHeight:  blockHeight,
		BlockHeader:  blockHeader,
		Transactions: transactions,
	}
}

// BlockHash returns the hash of this block's header
func (indexedBlock *IndexedBlock) BlockHash() chainhash.Hash {
	return indexedBlock.BlockHeader.BlockHash()
}
