package bitcoinspv

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	errHeaderInvalid   = errors.New("header is not valid")
	errParentNotFound  = errors.New("parent header cannot be found")
	errDuplicateHeader = errors.New("header was already submitted")
)

// unrecoverableErrors is a list of errors which are unsafe and should not be retried.
var unrecoverableErrors = []error{
	errHeaderInvalid,
	errParentNotFound,
}

// expectedErrors is a list of errors which can safely be ignored and should not be retried.
var expectedErrors = []error{
	errDuplicateHeader,
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

// RetryDo executes a func with retry
func RetryDo(sleep time.Duration, maxSleepDuration time.Duration, retryableFunc func() error) error {
	err := retryableFunc()
	if err == nil {
		return nil
	}

	if isUnrecoverableErr(err) {
		log.Err(err).Msg("Skip retry, error unrecoverable")
		return err
	}
	if isExpectedErr(err) {
		log.Err(err).Msg("Skip retry, error expected")
		return nil
	}

	// Add some randomness to prevent thrashing
	r, err := randDuration(int64(sleep))
	if err != nil {
		return err
	}
	sleep += r / 2

	if sleep > maxSleepDuration {
		log.Info().Msg("Retry timed out")
		return err
	}

	log.Debug().Err(err).Dur("sleep", sleep).Msg("Starting exponential backoff")
	time.Sleep(sleep)

	return RetryDo(2*sleep, maxSleepDuration, retryableFunc)
}

func randDuration(maxNumber int64) (time.Duration, error) {
	randNumber, err := rand.Int(rand.Reader, big.NewInt(maxNumber))
	if err != nil {
		return 0, err
	}
	return time.Duration(randNumber.Int64()), nil
}
