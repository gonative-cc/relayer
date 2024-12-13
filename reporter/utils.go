package reporter

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/reporter/types"
)

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// getHeaderMsgsToSubmit creates a set of MsgInsertHeaders messages corresponding to headers that
// should be submitted to Native light client from a given set of indexed blocks
func (r *Reporter) getHeaderMsgsToSubmit(
	ibs []*types.IndexedBlock,
) ([][]*wire.BlockHeader, error) {
	var (
		startPoint  = -1
		ibsToSubmit []*types.IndexedBlock
		err         error
	)

	// find the first header that is not contained in BBN header chain,
	// then submit since this header
	for i, _ := range ibs {
		// blockHash := header.BlockHash()
		// var res *btclctypes.QueryContainsBytesResponse
		var res bool
		err = RetryDo(r.retrySleepTime, r.maxRetrySleepTime, func() error {
			// TODO: placeholder true is used here for now
			// res, err = r.nativeClient.ContainsBTCBlock(&blockHash)
			res = true
			return err
		})
		if err != nil {
			return nil, err
		}
		if !res {
			startPoint = i
			break
		}
	}

	// all headers are duplicated, no need to submit
	if startPoint == -1 {
		r.logger.Info("All headers are duplicated, no need to submit")
		return nil, nil
	}

	// wrap the headers to MsgInsertHeaders msgs from the subset of indexed blocks
	ibsToSubmit = ibs[startPoint:]

	blockChunks := chunkBy(ibsToSubmit, int(r.Cfg.MaxHeadersInMsg))

	headerMsgsToSubmit := [][]*wire.BlockHeader{}

	for _, ibChunk := range blockChunks {
		msgInsertHeaders := types.NewMsgInsertHeaders(ibChunk)
		headerMsgsToSubmit = append(headerMsgsToSubmit, msgInsertHeaders)
	}

	return headerMsgsToSubmit, nil
}

func (r *Reporter) submitHeaderMsgs(msg []*wire.BlockHeader) error {
	// submit the headers
	err := RetryDo(r.retrySleepTime, r.maxRetrySleepTime, func() error {
		err := r.nativeClient.InsertHeaders(msg)
		if err != nil {
			return err
		}
		r.logger.Infof(
			"Successfully submitted %d headers to Babylon", len(msg),
		)
		return nil
	})
	if err != nil {
		r.metrics.FailedHeadersCounter.Add(float64(len(msg)))
		return fmt.Errorf("failed to submit headers: %w", err)
	}

	// update metrics
	r.metrics.SuccessfulHeadersCounter.Add(float64(len(msg)))
	r.metrics.SecondsSinceLastHeaderGauge.Set(0)
	for _, header := range msg {
		r.metrics.NewReportedHeaderGaugeVec.WithLabelValues(
			header.BlockHash().String(),
		).SetToCurrentTime()
	}

	return err
}

// ProcessHeaders extracts and reports headers from a list of blocks
// It returns the number of headers that need to be reported (after deduplication)
func (r *Reporter) ProcessHeaders(ibs []*types.IndexedBlock) (int, error) {
	// get a list of MsgInsertHeader msgs with headers to be submitted
	headerMsgsToSubmit, err := r.getHeaderMsgsToSubmit(ibs)
	if err != nil {
		return 0, fmt.Errorf("failed to find headers to submit: %w", err)
	}
	// skip if no header to submit
	if len(headerMsgsToSubmit) == 0 {
		r.logger.Info("No new headers to submit")
		return 0, nil
	}

	var numSubmitted int
	// submit each chunk of headers
	for _, msgs := range headerMsgsToSubmit {
		if err := r.submitHeaderMsgs(msgs); err != nil {
			return 0, fmt.Errorf("failed to submit headers: %w", err)
		}
		numSubmitted += len(msgs)
	}

	return numSubmitted, err
}

// push msg to channel c, or quit if quit channel is closed
func PushOrQuit[T any](c chan<- T, msg T, quit <-chan struct{}) {
	select {
	case c <- msg:
	case <-quit:
	}
}
