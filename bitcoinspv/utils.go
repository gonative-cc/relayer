package bitcoinspv

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

func chunkBy[T any](items []T, chunkSize int) [][]T {
	if len(items) == 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(items)+chunkSize-1)/chunkSize)
	for i := 0; i < len(items); i += chunkSize {
		end := i + chunkSize
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// getHeaderMsgsToSubmit creates a set of MsgInsertHeaders messages corresponding to headers that
// should be submitted to Native light client from a given set of indexed blocks
func (r *Relayer) getHeaderMsgsToSubmit(
	indexed_blocks []*types.IndexedBlock,
) ([][]*wire.BlockHeader, error) {
	startPoint, err := r.findFirstNewHeader(indexed_blocks)
	if err != nil {
		return nil, err
	}

	// all headers are duplicated, no need to submit
	if startPoint == -1 {
		r.logger.Info("All headers are duplicated, no need to submit")
		return nil, nil
	}

	// Get subset of blocks starting from first new header
	blocksToSubmit := indexed_blocks[startPoint:]

	// Split into chunks and convert to header messages
	return r.createHeaderMessages(blocksToSubmit), nil
}

// findFirstNewHeader finds the index of the first header not in the Native chain
func (r *Relayer) findFirstNewHeader(indexed_blocks []*types.IndexedBlock) (int, error) {
	for i, header := range indexed_blocks {
		blockHash := header.BlockHash()
		var res bool
		var err error
		err = RetryDo(r.retrySleepDuration, r.maxRetrySleepDuration, func() error {
			res, err = r.nativeClient.ContainsBTCBlock(&blockHash)
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
func (r *Relayer) createHeaderMessages(indexed_blocks []*types.IndexedBlock) [][]*wire.BlockHeader {
	blockChunks := chunkBy(indexed_blocks, int(r.Config.MaxHeadersInMsg))
	headerMsgs := make([][]*wire.BlockHeader, 0, len(blockChunks))

	for _, chunk := range blockChunks {
		headerMsgs = append(headerMsgs, types.NewMsgInsertHeaders(chunk))
	}

	return headerMsgs
}

func (r *Relayer) submitHeaderMsgs(msg []*wire.BlockHeader) error {
	// submit the headers
	err := RetryDo(r.retrySleepDuration, r.maxRetrySleepDuration, func() error {
		if err := r.nativeClient.InsertHeaders(msg); err != nil {
			return err
		}
		r.logger.Infof(
			"Successfully submitted %d headers to light client", len(msg),
		)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to submit headers: %w", err)
	}

	return nil
}

// ProcessHeaders extracts and reports headers from a list of blocks
// It returns the number of headers that need to be reported (after deduplication)
func (r *Relayer) ProcessHeaders(indexed_blocks []*types.IndexedBlock) (int, error) {
	headerMsgsToSubmit, err := r.getHeaderMsgsToSubmit(indexed_blocks)
	if err != nil {
		return 0, fmt.Errorf("failed to find headers to submit: %w", err)
	}
	if len(headerMsgsToSubmit) == 0 {
		r.logger.Info("No new headers to submit")
	}

	numSubmitted := 0
	for _, msgs := range headerMsgsToSubmit {
		if err := r.submitHeaderMsgs(msgs); err != nil {
			return 0, fmt.Errorf("failed to submit headers: %w", err)
		}
		numSubmitted += len(msgs)
	}

	return numSubmitted, nil
}

func (r *Relayer) extractAndSubmitTransactions(ib *types.IndexedBlock) (int, error) {
	numSubmittedTxs := 0

	for txIdx, tx := range ib.Txs {
		if err := r.submitTransaction(ib, txIdx, tx); err != nil {
			return numSubmittedTxs, err
		}
		numSubmittedTxs++
	}

	return numSubmittedTxs, nil
}

func (r *Relayer) submitTransaction(ib *types.IndexedBlock, txIdx int, tx *btcutil.Tx) error {
	if tx == nil {
		r.logger.Warnf("Found a nil tx in block %v", ib.BlockHash())
		return nil
	}

	// construct spv proof from tx
	//nolint:gosec
	proof, err := ib.GenSPVProof(uint32(txIdx)) // Ignore G115, txIdx always >= 0
	if err != nil {
		r.logger.Errorf("Failed to construct spv proof from tx %v: %v", tx.Hash(), err)
		return err
	}

	// wrap to MsgSpvProof
	msgSpvProof := proof.ToMsgSpvProof(tx.MsgTx().TxID(), tx.Hash())

	// submit the checkpoint to light client
	res, err := r.nativeClient.VerifySPV(&msgSpvProof)
	if err != nil {
		r.logger.Errorf("Failed to submit MsgInsertBTCSpvProof with error %v", err)
		return err
	}

	r.logger.Infof("Successfully submitted MsgInsertBTCSpvProof with response %d", res)

	return nil
}

// ProcessTransactions tries to extract valid transactions from a list of blocks
// It returns the number of valid transactions segments, and the number of valid transactions
func (r *Relayer) ProcessTransactions(indexed_blocks []*types.IndexedBlock) (int, error) {
	var totalTxs int

	// process transactions from each block
	for _, block := range indexed_blocks {
		blockTxs, err := r.extractAndSubmitTransactions(block)
		if err != nil {
			if totalTxs > 0 {
				r.logger.Infof("Submitted %d transactions", totalTxs)
			}
			return totalTxs, fmt.Errorf(
				"failed to extract transactions from block %v: %w", block.BlockHash(), err,
			)
		}
		totalTxs += blockTxs
	}

	// log total transactions processed if any
	if totalTxs > 0 {
		r.logger.Infof("Submitted %d transactions", totalTxs)
	}

	return totalTxs, nil
}

// push msg to channel c, or quit if quit channel is closed
func PushOrQuit[T any](c chan<- T, msg T, quit <-chan struct{}) {
	select {
	case c <- msg:
	case <-quit:
	}
}
