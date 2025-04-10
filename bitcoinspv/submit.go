package bitcoinspv

import (
	"context"
	"fmt"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// createChunks takes a set of indexed blocks and breaks them into chunks of headers to be sent
// to the light client.
func (r *Relayer) createChunks(ctx context.Context, indexedBlocks []*types.IndexedBlock,
) ([]Chunk, error) {
	startPoint, err := r.findFirstNewHeader(ctx, indexedBlocks)
	if err != nil {
		return nil, err
	}

	if startPoint == -1 {
		r.logger.Info().Msg("All headers are duplicated, no need to submit")
		return nil, nil
	}

	blocksToSubmit := indexedBlocks[startPoint:]
	blockChunks := breakIntoChunks(blocksToSubmit, int(r.Config.HeadersChunkSize))
	return blockChunks, nil
}

// findFirstNewHeader finds the index of the first header not in the light client.
func (r *Relayer) findFirstNewHeader(ctx context.Context, indexedBlocks []*types.IndexedBlock) (int, error) {
	for i, header := range indexedBlocks {
		blockHash := header.BlockHash()
		var res bool
		var err error
		err = RetryDo(r.logger, r.Config.RetrySleepDuration, r.Config.MaxRetrySleepDuration, func() error {
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

func (r *Relayer) submitHeaderMessages(ctx context.Context, chunk Chunk) error {
	err := RetryDo(r.logger, r.Config.RetrySleepDuration, r.Config.MaxRetrySleepDuration, func() error {
		if err := r.lcClient.InsertHeaders(ctx, chunk.Headers); err != nil {
			return err
		}
		hs := chunk.Headers
		firstHash := hs[0].BlockHash().String()
		var headersStr string
		if len(hs) > 1 {
			lastHash := hs[len(hs)-1].BlockHash().String()
			headersStr = fmt.Sprint("headers=[", firstHash, "...", lastHash, "]",
				" heights=[", chunk.From, "...", chunk.To, "]")
		} else {
			headersStr = fmt.Sprint("header=", firstHash, " height=", chunk.From)
		}
		r.logger.Info().Msgf("Submitted %d %s to light client", len(hs), headersStr)
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
	chunks, err := r.createChunks(ctx, indexedBlocks)
	if err != nil {
		return 0, fmt.Errorf("failed to find headers to submit: %w", err)
	}
	if len(chunks) == 0 {
		r.logger.Info().Msg("No new headers to submit")
	}

	headersSubmitted := 0
	for _, chunk := range chunks {
		if err := r.submitHeaderMessages(ctx, chunk); err != nil {
			return 0, fmt.Errorf("failed to submit headers: %w", err)
		}
		headersSubmitted += len(chunk.Headers)
	}

	return headersSubmitted, nil
}
