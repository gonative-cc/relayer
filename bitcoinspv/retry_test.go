package bitcoinspv

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const (
	retryInterval = 1 * time.Second
	retryTimeout  = 1 * time.Minute
)

func TestUnrecoverableError(t *testing.T) {
	err := RetryDo(zerolog.Logger{}, retryInterval, retryTimeout, func() error {
		return unrecoverableErrors[0]
	})
	require.Error(t, err)
}

func TestExpectedError(t *testing.T) {
	err := RetryDo(zerolog.Logger{}, retryInterval, retryTimeout, func() error {
		return expectedErrors[0]
	})
	require.NoError(t, err)
}
