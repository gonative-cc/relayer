package bitcoinspv

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

func breakIntoChunks[T any](v []T, chunkSize int) [][]T {
	if len(v) == 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(v)+chunkSize-1)/chunkSize)
	for i := 0; i < len(v); i += chunkSize {
		end := i + chunkSize
		if end > len(v) {
			end = len(v)
		}
		chunks = append(chunks, v[i:end])
	}
	return chunks
}

// getHeaderMessages takes a set of indexed blocks and generates MsgInsertHeaders messages
// containing block headers that need to be sent to the light client.
func (r *Relayer) getHeaderMessages(ctx context.Context,
	indexedBlocks []*types.IndexedBlock,
) ([][]wire.BlockHeader, error) {
	startPoint, err := r.findFirstNewHeader(ctx, indexedBlocks)
	if err != nil {
		return nil, err
	}

	if startPoint == -1 {
		r.logger.Info("All headers are duplicated, no need to submit")
		return nil, nil
	}

	// Get subset of blocks starting from first new header
	blocksToSubmit := indexedBlocks[startPoint:]

	// Split into chunks and convert to header messages
	return r.createHeaderMessages(blocksToSubmit), nil
}

// findFirstNewHeader finds the index of the first header not in the light client.
func (r *Relayer) findFirstNewHeader(ctx context.Context, indexedBlocks []*types.IndexedBlock) (int, error) {
	for i, header := range indexedBlocks {
		blockHash := header.BlockHash()
		var res bool
		var err error
		err = RetryDo(r.Config.RetrySleepDuration, r.Config.MaxRetrySleepDuration, func() error {
			res, err = r.lcClient.ContainsBlock(ctx, blockHash)
			return err
		})
		if err != nil {
			return -1, err
		}
		if !res {
			return i, nil
		}
	}
	return -1, nil
}

// createHeaderMessages splits blocks into chunks and creates header messages
func (r *Relayer) createHeaderMessages(indexedBlocks []*types.IndexedBlock) [][]wire.BlockHeader {
	blockChunks := breakIntoChunks(indexedBlocks, int(r.Config.HeadersChunkSize))
	headerMsgs := make([][]wire.BlockHeader, 0, len(blockChunks))

	for _, chunk := range blockChunks {
		headerMsgs = append(headerMsgs, types.NewMsgInsertHeaders(chunk))
	}

	return headerMsgs
}

func (r *Relayer) submitHeaderMessages(ctx context.Context, headers []wire.BlockHeader) error {
	err := RetryDo(r.Config.RetrySleepDuration, r.Config.MaxRetrySleepDuration, func() error {
		if err := r.lcClient.InsertHeaders(ctx, headers); err != nil {
			return err
		}
		first := headers[0].BlockHash().String()
		last := headers[len(headers)-1].BlockHash().String()
		r.logger.Infof(
			"Submitted %d headers [%s ... %s] to light client", len(headers), first, last)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to submit headers: %w", err)
	}
	return nil
}

// ProcessHeaders takes a list of blocks, extracts their headers
// and submits them to the light client.
// Returns the count of unique headers that were submitted.
func (r *Relayer) ProcessHeaders(ctx context.Context, indexedBlocks []*types.IndexedBlock) (int, error) {
	headerMessages, err := r.getHeaderMessages(ctx, indexedBlocks)
	if err != nil {
		return 0, fmt.Errorf("failed to find headers to submit: %w", err)
	}
	if len(headerMessages) == 0 {
		r.logger.Info("No new headers to submit")
	}

	headersSubmitted := 0
	for _, msgs := range headerMessages {
		if err := r.submitHeaderMessages(ctx, msgs); err != nil {
			return 0, fmt.Errorf("failed to submit headers: %w", err)
		}
		headersSubmitted += len(msgs)
	}

	return headersSubmitted, nil
}
