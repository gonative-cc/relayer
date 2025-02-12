package types

import "github.com/btcsuite/btcd/wire"

type MsgInsertHeaders struct {
	Headers []BTCHeaderBytes
}

// NewMsgInsertHeaders converts a slice of IndexedBlock to wire.BlockHeader slice
func NewMsgInsertHeaders(blocks []*IndexedBlock) []*wire.BlockHeader {
	result := make([]*wire.BlockHeader, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, block.BlockHeader)
	}
	return result
}
