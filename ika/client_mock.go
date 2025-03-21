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

// SignReq is a mock implementation that returns custom values based on test setup.
func (m *MockClient) SignReq(
	ctx context.Context,
	dwalletCapID string,
	signMessagesID string,
	messages [][]byte,
) (string, error) {
	returns := m.Called(ctx, dwalletCapID, signMessagesID, messages)
	return returns.String(0), returns.Error(1)
}

// QuerySign queries signatures from Ika
func (m *MockClient) QuerySign() {}
