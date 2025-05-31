package bitcoinspv

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	sui_errors "github.com/gonative-cc/relayer/bitcoinspv/clients/sui"
	"github.com/rs/zerolog"
)

// ErrRetryTimeout indicates that the operation failed after all retry attempts.
var ErrRetryTimeout = errors.New("retry timed out after maximum duration")

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
	case categoryUnknown:
		logger.Err(err).Msg("Skip retry, error classification failed or unknown type")
	case categoryRetryable:
		logger.Debug().Err(err).Msg("Retryable error, trying to repeat the request")
		// Add some randomness to prevent thrashing
		jitter, randErr := randDuration(int64(sleep))
		if randErr != nil {
			logger.Err(randErr).Msg("Failed to generate random jitter during retry")
			return randErr
		}
		sleep += jitter / 2

		if sleep > maxSleepDuration {
			logger.Err(err).Dur("sleep_limit", maxSleepDuration).Msg("Retry timed out")
			return fmt.Errorf("%w: last error was %w", ErrRetryTimeout, err)
		}

		logger.Debug().Dur("sleep", sleep).Msg("Exponential backoff")
		time.Sleep(sleep)

		return RetryDo(logger, 2*sleep, maxSleepDuration, retryableFunc)
	default:
		logger.Err(err).Int("category", int(category)).Msg("Unhandled error category in RetryDo")
		return err
	}
	return err
}

func randDuration(maxNumber int64) (time.Duration, error) {
	randNumber, err := rand.Int(rand.Reader, big.NewInt(maxNumber))
	if err != nil {
		return 0, err
	}
	return time.Duration(randNumber.Int64()), nil
}
