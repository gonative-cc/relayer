package bitcoinspv

import (
	"crypto/rand"
	"errors"
	"math/big"
	"os"
	"time"

	"cosmossdk.io/log"
)

var ErrInvalidHeader = errors.New("invalid header")
var ErrDuplicatedSubmission = errors.New("duplicated header submitted")

// TODO add log formatters
var logger = log.NewLogger(os.Stdout)

// unrecoverableErrors is a list of errors which are unsafe and should not be retried.
var unrecoverableErrors = []error{
	ErrInvalidHeader,
	// populate list of errors
	// btclctypes.ErrHeaderParentDoesNotExist,
	// btclctypes.ErrChainWithNotEnoughWork,
	// btcctypes.ErrProvidedHeaderDoesNotHaveAncestor,
	// btcctypes.ErrNoCheckpointsForPreviousEpoch,
	// btcctypes.ErrInvalidCheckpointProof,
	// TODO Add more errors here
}

// expectedErrors is a list of errors which can safely be ignored and should not be retried.
var expectedErrors = []error{
	ErrDuplicatedSubmission,
	// btcctypes.ErrInvalidHeader,
	// TODO Add more errors here
}

func isUnrecoverableErr(err error) bool {
	for _, e := range unrecoverableErrors {
		if errors.Is(err, e) {
			return true
		}
	}

	return false
}

func isExpectedErr(err error) bool {
	for _, e := range expectedErrors {
		if errors.Is(err, e) {
			return true
		}
	}

	return false
}

// Do executes a func with retry
func RetryDo(sleep time.Duration, maxSleepTime time.Duration, retryableFunc func() error) error {
	if err := retryableFunc(); err != nil {
		if isUnrecoverableErr(err) {
			logger.Error("Skip retry, error unrecoverable", "err", err)
			return err
		}

		if isExpectedErr(err) {
			logger.Error("Skip retry, error expected", "err", err)
			return nil
		}

		// Add some randomness to prevent thrashing
		jitter, err := randDuration(int64(sleep))
		if err != nil {
			return err
		}
		sleep += jitter / 2

		if sleep > maxSleepTime {
			logger.Info("retry timed out")
			return err
		}

		logger.Info("starting exponential backoff", "sleep", sleep, "err", err)
		time.Sleep(sleep)

		return RetryDo(2*sleep, maxSleepTime, retryableFunc)
	}
	return nil
}

func randDuration(maxNumber int64) (dur time.Duration, err error) {
	randNumber, err := rand.Int(rand.Reader, big.NewInt(maxNumber))
	if err != nil {
		return dur, err
	}
	return time.Duration(randNumber.Int64()), nil
}
