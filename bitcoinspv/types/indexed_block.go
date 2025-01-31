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
	Height int64
	Header *wire.BlockHeader
	Txs    []*btcutil.Tx
}

func NewIndexedBlock(height int64, header *wire.BlockHeader, txs []*btcutil.Tx) *IndexedBlock {
	return &IndexedBlock{
		Height: height,
		Header: header,
		Txs:    txs,
	}
}

func NewIndexedBlockFromMsgBlock(height int64, block *wire.MsgBlock) *IndexedBlock {
	return &IndexedBlock{
		Height: height,
		Header: &block.Header,
		Txs:    GetWrappedTxs(block),
	}
}

func (ib *IndexedBlock) MsgBlock() *wire.MsgBlock {
	msgTxs := make([]*wire.MsgTx, 0, len(ib.Txs))
	for _, tx := range ib.Txs {
		msgTxs = append(msgTxs, tx.MsgTx())
	}

	return &wire.MsgBlock{
		Header:       *ib.Header,
		Transactions: msgTxs,
	}
}

func (ib *IndexedBlock) BlockHash() chainhash.Hash {
	return ib.Header.BlockHash()
}

// GenSPVProof generates a Merkle proof for the transaction at the given index
func (ib *IndexedBlock) GenSPVProof(txIdx uint32) (*BTCSpvProof, error) {
	if int(txIdx) >= len(ib.Txs) {
		return nil, fmt.Errorf(
			"transaction index is out of scope: idx=%d, len(Txs)=%d",
			txIdx, len(ib.Txs),
		)
	}

	headerBytes := NewBTCHeaderBytesFromBlockHeader(ib.Header)
	txsBytes := ib.serializeTransactions()
	if len(txsBytes) == 0 {
		return nil, fmt.Errorf("failed to serialize transactions")
	}

	return SpvProofFromHeaderAndTransactions(&headerBytes, txsBytes, txIdx)
}

// serializeTransactions converts all transactions to byte slices
func (ib *IndexedBlock) serializeTransactions() [][]byte {
	txsBytes := make([][]byte, 0, len(ib.Txs))
	for _, tx := range ib.Txs {
		var txBuf bytes.Buffer
		if err := tx.MsgTx().Serialize(&txBuf); err != nil {
			return nil
		}
		txsBytes = append(txsBytes, txBuf.Bytes())
	}
	return txsBytes
}
