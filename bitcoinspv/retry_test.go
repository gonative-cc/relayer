package bitcoinspv

import (
	// Added imports
	"errors"
	"fmt"
	"testing"
	"time"

	sui_errors "github.com/gonative-cc/relayer/bitcoinspv/clients/sui"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const (
	retryInterval = 2 * time.Millisecond
	retryTimeout  = 50 * time.Millisecond
)

func TestNonRecoverableError(t *testing.T) {
	t.Run("MoveAbort", func(t *testing.T) {
		simulatedMoveAbortErr := fmt.Errorf("%w: function 'test_abort' status: failure, error: MoveAbort(MoveLocation { module: ..., name: ... }, 1234567890) in command 0",
			sui_errors.ErrSuiTransactionFailed)

		testRetryDo(t, simulatedMoveAbortErr)
	})

	t.Run("OutOfGas", func(t *testing.T) {
		simulatedOutOfGasErr := fmt.Errorf("%w: function 'test_gas' status: failure, error: OutOfGas",
			sui_errors.ErrSuiTransactionFailed)

		testRetryDo(t, simulatedOutOfGasErr)
	})

	t.Run("OtherExecutionFailure", func(t *testing.T) {
		simulatedOtherFailureErr := fmt.Errorf("%w: function 'test_other' status: failure, error: SomeOtherExecutionError",
			sui_errors.ErrSuiTransactionFailed)

		testRetryDo(t, simulatedOtherFailureErr)
	})
}

func TestRetryableError(t *testing.T) {
	// An error that does not wrap ErrSuiTransactionFailed
	simulatedRetryableErr := errors.New("temporary network issue")

	maxCalls := 3
	callCount := 0
	err := RetryDo(zerolog.Nop(), retryInterval, retryTimeout, func() error {
		callCount++
		if callCount < maxCalls {
			return simulatedRetryableErr
		}
		// Success on the last call
		return nil
	})

	assert.NoError(t, err, "RetryDo should eventually succeed")
	assert.Equal(t, maxCalls, callCount, "Function should be called multiple times")
}

func TestRetryableErrorTimeout(t *testing.T) {
	simulatedRetryableErr := errors.New("temporary network issue")
	err := RetryDo(zerolog.Nop(), retryInterval, 4*time.Millisecond, func() error {
		return simulatedRetryableErr
	})

	assert.ErrorIs(t, err, ErrRetryTimeout)
	assert.ErrorIs(t, err, simulatedRetryableErr)
}

func testRetryDo(t *testing.T, mockErr error) {
	callCount := 0
	err := RetryDo(zerolog.Nop(), retryInterval, retryTimeout, func() error {
		callCount++
		return mockErr
	})
	assert.Error(t, err, "RetryDo should return an error")
	assert.Equal(t, 1, callCount, "Function should be called only once")
}
