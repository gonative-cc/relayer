package types

import "github.com/btcsuite/btcd/wire"

type MsgInsertHeaders struct {
	Headers []BTCHeaderBytes
}

// NewMsgInsertHeaders converts a slice of IndexedBlock to wire.BlockHeader
func NewMsgInsertHeaders(headers []*IndexedBlock) []*wire.BlockHeader {
	headerBytes := make([]*wire.BlockHeader, len(headers))
	for i, header := range headers {
		headerBytes[i] = header.Header
	}

	return headerBytes
}
