package bitcoinspv

import (
	"github.com/btcsuite/btcd/wire"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

func breakIntoChunks(blocks []*types.IndexedBlock, chunkSize int) []Chunk {
	if chunkSize < 1 {
		return nil
	}
	if len(blocks) == 0 {
		return nil
	}

	chunks := make([]Chunk, 0, (len(blocks)+chunkSize-1)/chunkSize)
	for i := 0; i < len(blocks); i += chunkSize {
		end := i + chunkSize
		if end > len(blocks) {
			end = len(blocks)
		}
		chunk := Chunk{
			From:    blocks[i].BlockHeight,
			To:      blocks[end-1].BlockHeight,
			Headers: toBlockHeaders(blocks[i:end]),
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func toBlockHeaders(blocks []*types.IndexedBlock) []wire.BlockHeader {
	headers := make([]wire.BlockHeader, 0, len(blocks))
	for _, block := range blocks {
		headers = append(headers, block.RawMsgBlock.Header)
	}
	return headers
}

// Chunk represents a batch of block headers to be sent to the light client.
type Chunk struct {
	Headers []wire.BlockHeader
	From    int64
	To      int64
}
