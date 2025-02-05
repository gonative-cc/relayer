package bitcoinspv

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// getHeaderMsgsToSubmit creates a set of MsgInsertHeaders messages corresponding to headers that
// should be submitted to Native light client from a given set of indexed blocks
func (r *Relayer) getHeaderMsgsToSubmit(
	ibs []*types.IndexedBlock,
) ([][]*wire.BlockHeader, error) {
	startPoint, err := r.findFirstNewHeader(ibs)
	if err != nil {
		return nil, err
	}

	// all headers are duplicated, no need to submit
	if startPoint == -1 {
		r.logger.Info("all headers are duplicated, no need to submit")
		return nil, nil
	}

	// Get subset of blocks starting from first new header
	ibsToSubmit := ibs[startPoint:]

	// Split into chunks and convert to header messages
	return r.createHeaderMessages(ibsToSubmit), nil
}

// findFirstNewHeader finds the index of the first header not in the Native chain
func (r *Relayer) findFirstNewHeader(ibs []*types.IndexedBlock) (int, error) {
	for i, header := range ibs {
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
func (r *Relayer) createHeaderMessages(ibs []*types.IndexedBlock) [][]*wire.BlockHeader {
	blockChunks := chunkBy(ibs, int(r.Config.MaxHeadersInMsg))
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
			"successfully submitted %d headers to light client", len(msg),
		)
		return nil
	})
	if err != nil {
		r.metrics.FailedHeadersCounter.Add(float64(len(msg)))
		return fmt.Errorf("failed to submit headers: %w", err)
	}

	// update metrics
	r.updateHeaderMetrics(msg)

	return nil
}

func (r *Relayer) updateHeaderMetrics(headers []*wire.BlockHeader) {
	r.metrics.SuccessfulHeadersCounter.Add(float64(len(headers)))
	r.metrics.SecondsSinceLastHeaderGauge.Set(0)
	for _, header := range headers {
		r.metrics.NewReportedHeaderGaugeVec.WithLabelValues(
			header.BlockHash().String(),
		).SetToCurrentTime()
	}
}

// ProcessHeaders extracts and reports headers from a list of blocks
// It returns the number of headers that need to be reported (after deduplication)
func (r *Relayer) ProcessHeaders(ibs []*types.IndexedBlock) (int, error) {
	// get a list of MsgInsertHeader msgs with headers to be submitted
	headerMsgsToSubmit, err := r.getHeaderMsgsToSubmit(ibs)
	if err != nil {
		return 0, fmt.Errorf("failed to find headers to submit: %w", err)
	}
	// skip if no header to submit
	if len(headerMsgsToSubmit) == 0 {
		r.logger.Info("no new headers to submit")
		return 0, nil
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
		r.logger.Warnf("found a nil tx in block %v", ib.BlockHash())
		return nil
	}

	// construct spv proof from tx
	//nolint:gosec
	proof, err := ib.GenSPVProof(uint32(txIdx)) // Ignore G115, txIdx always >= 0
	if err != nil {
		r.logger.Errorf("failed to construct spv proof from tx %v: %v", tx.Hash(), err)
		return err
	}

	// wrap to MsgSpvProof
	msgSpvProof := proof.ToMsgSpvProof(tx.MsgTx().TxID(), tx.Hash())

	// submit the checkpoint to light client
	res, err := r.nativeClient.VerifySPV(&msgSpvProof)
	if err != nil {
		r.logger.Errorf("failed to submit MsgInsertBTCSpvProof with error %v", err)
		r.metrics.FailedCheckpointsCounter.Inc()
		return err
	}

	r.logger.Infof("successfully submitted MsgInsertBTCSpvProof with response %d", res)

	// metrics sent to prometheus instance
	r.metrics.SuccessfulCheckpointsCounter.Inc()
	r.metrics.SecondsSinceLastCheckpointGauge.Set(0)
	// tx1Block := ckpt.Segments[0].AssocBlock
	// tx2Block := ckpt.Segments[1].AssocBlock
	// r.metrics.NewReportedCheckpointGaugeVec.WithLabelValues(
	// 	strconv.Itoa(int(ckpt.Epoch)),
	// 	strconv.Itoa(int(tx1Block.Height)),
	// 	tx1Block.Txs[ckpt.Segments[0].TxIdx].Hash().String(),
	// 	tx2Block.Txs[ckpt.Segments[1].TxIdx].Hash().String(),
	// ).SetToCurrentTime()

	return nil
}

// ProcessTransactions tries to extract valid transactions from a list of blocks
// It returns the number of valid transactions segments, and the number of valid transactions
func (r *Relayer) ProcessTransactions(ibs []*types.IndexedBlock) (int, error) {
	var totalTxs int

	// process transactions from each block
	for _, block := range ibs {
		blockTxs, err := r.extractAndSubmitTransactions(block)
		if err != nil {
			if totalTxs > 0 {
				r.logger.Infof("submitted %d transactions", totalTxs)
			}
			return totalTxs, fmt.Errorf(
				"failed to extract transactions from block %v: %w", block.BlockHash(), err,
			)
		}
		totalTxs += blockTxs
	}

	// log total transactions processed if any
	if totalTxs > 0 {
		r.logger.Infof("submitted %d transactions", totalTxs)
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
