package remote2ika

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface for testing.
type MockClient struct {
	mock.Mock
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
