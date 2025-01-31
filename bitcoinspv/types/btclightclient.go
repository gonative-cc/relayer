package types

import "github.com/btcsuite/btcd/wire"

func NewMsgInsertHeaders(
	headers []*IndexedBlock,
) []*wire.BlockHeader {
	headerBytes := make([]*wire.BlockHeader, len(headers))
	for i, h := range headers {
		headerBytes[i] = h.Header
	}

	return headerBytes
}
