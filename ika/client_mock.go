package ika

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface for testing.
type MockClient struct {
	mock.Mock
}

// NewMockClient creates a new mock Client instance with predefined ApproveAndSign behavior.
func NewMockClient() *MockClient {
	mockClient := new(MockClient)
	mockClient.On("ApproveAndSign", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		// predefined signature and transaction digest.
		Return([][]byte{{0x01, 0x02, 0x03}}, "txDigest", nil)
	return mockClient
}

// UpdateLC is a mock implementation used in tests.
func (m *MockClient) UpdateLC(
	_ context.Context,
	_ *tmtypes.LightBlock,
	_ zerolog.Logger,
) (models.SuiTransactionBlockResponse, error) {
	return models.SuiTransactionBlockResponse{}, nil
}

// ApproveAndSign is a mock implementation that returns custom values based on test setup.
func (m *MockClient) ApproveAndSign(
	ctx context.Context,
	dwalletCapID string,
	signMessagesID string,
	messages [][]byte,
) ([][]byte, string, error) {
	returns := m.Called(ctx, dwalletCapID, signMessagesID, messages)
	var signatures [][]byte
	if returns.Get(0) == nil {
		// When passing nil as the first arguemnt testify complains about nil not being of type [][]byte,
		// thats where make([][]byte, 0) come from.
		signatures = make([][]byte, 0)
	} else {
		signatures = returns.Get(0).([][]byte)
	}
	return signatures, returns.String(1), returns.Error(2)
}
