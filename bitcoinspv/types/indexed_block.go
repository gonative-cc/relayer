package types

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// IndexedBlock represents a BTC block with additional metadata including block height
// and transaction details needed for Merkle proof generation
type IndexedBlock struct {
	BlockHeight  int64
	BlockHeader  *wire.BlockHeader
	Transactions []*btcutil.Tx
}

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

func (indexedBlock *IndexedBlock) BlockHash() chainhash.Hash {
	return indexedBlock.BlockHeader.BlockHash()
}

// GenerateProof creates a Merkle proof for the specified transaction
// using its index in the block
func (indexedBlock *IndexedBlock) GenerateProof(txIdx uint32) (*BTCSpvProof, error) {
	if err := indexedBlock.validateTxIndex(txIdx); err != nil {
		return nil, err
	}

	headerBytes := NewBTCHeaderBytesFromBlockHeader(indexedBlock.BlockHeader)
	txsBytes := indexedBlock.serializeTransactions()
	if len(txsBytes) == 0 {
		return nil, fmt.Errorf("failed to serialize transactions")
	}

	return SpvProofFromHeaderAndTransactions(&headerBytes, txsBytes, txIdx)
}

func (indexedBlock *IndexedBlock) validateTxIndex(txIdx uint32) error {
	if int(txIdx) >= len(indexedBlock.Transactions) {
		return fmt.Errorf(
			"transaction index is out of scope: idx=%d, len(Txs)=%d",
			txIdx, len(indexedBlock.Transactions),
		)
	}
	return nil
}

// serializeTransactions converts all transactions to byte slices
func (indexedBlock *IndexedBlock) serializeTransactions() [][]byte {
	txsBytes := make([][]byte, 0, len(indexedBlock.Transactions))
	for _, tx := range indexedBlock.Transactions {
		var txBuf bytes.Buffer
		if err := tx.MsgTx().Serialize(&txBuf); err != nil {
			return nil
		}
		txsBytes = append(txsBytes, txBuf.Bytes())
	}
	return txsBytes
}
