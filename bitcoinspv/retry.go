package bitcoinspv

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	sui_errors "github.com/gonative-cc/relayer/bitcoinspv/clients/sui"
	"github.com/rs/zerolog"
)

// ErrorCategory classifies errors for retry logic.
type ErrorCategory int

const (
	categoryRetryable      ErrorCategory = iota // Network/infra errors
	categoryNonRecoverable                      // Stop retrying (MoveAbort, OutOfGas)
	categoryUnknown
)

func classifyError(logger zerolog.Logger, err error) ErrorCategory {
	if !errors.Is(err, sui_errors.ErrSuiTransactionFailed) {
		return categoryRetryable
	}

	// ELSE it is `ErrSuiTransactionFailed``, meaning execution completed but failed.
	if strings.Contains(err.Error(), "MoveAbort(") {
		return categoryNonRecoverable
	}

	if strings.Contains(err.Error(), "OutOfGas") {
		return categoryNonRecoverable
	}

	// IF ErrSuiTransactionFailed occurred, and the status was 'failure',
	// but we didn't specifically identify MoveAbort/OutOfGas,
	// it's still an execution failure. Treat as NonRetryable.
	logger.Warn().Msgf("Unidentified execution failure status within ErrSuiTransactionFailed: %s", err.Error())
	return categoryNonRecoverable
}

// RetryDo executes a func with retry
func RetryDo(
	logger zerolog.Logger,
	sleep time.Duration,
	maxSleepDuration time.Duration,
	retryableFunc func() error,
) error {
	err := retryableFunc()
	if err == nil {
		return nil
	}

	category := classifyError(logger, err)

	switch category {
	case categoryNonRecoverable:
		logger.Warn().Err(err).Msg("Skip retry, error classified as non-retryable")
		return err
	case categoryUnknown:
		logger.Error().Err(err).Msg("Skip retry, error classification failed or unknown type")
		return err
	case categoryRetryable:
		// Add some randomness to prevent thrashing
		jitter, randErr := randDuration(int64(sleep))
		if randErr != nil {
			logger.Error().Err(randErr).Msg("Failed to generate random jitter during retry")
			return randErr
		}
		sleep += jitter / 2

		if sleep > maxSleepDuration {
			logger.Err(randErr).Dur("sleep_limit", maxSleepDuration).Msg("Retry timed out")
			return randErr
		}

		logger.Debug().Err(randErr).Dur("sleep", sleep).Msg("Starting exponential backoff")
		time.Sleep(sleep)

		return RetryDo(logger, 2*sleep, maxSleepDuration, retryableFunc)
	default:
		logger.Error().Err(err).Int("category", int(category)).Msg("Unhandled error category in RetryDo")
		return err
	}
}

func randDuration(maxNumber int64) (time.Duration, error) {
	randNumber, err := rand.Int(rand.Reader, big.NewInt(maxNumber))
	if err != nil {
		return 0, err
	}
	return time.Duration(randNumber.Int64()), nil
}
