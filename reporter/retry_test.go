package reporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnrecoverableError(t *testing.T) {
	err := RetryDo(1*time.Second, 1*time.Minute, func() error {
		return unrecoverableErrors[0]
	})
	require.Error(t, err)
}

func TestExpectedError(t *testing.T) {
	err := RetryDo(1*time.Second, 1*time.Minute, func() error {
		return expectedErrors[0]
	})
	require.NoError(t, err)
}
