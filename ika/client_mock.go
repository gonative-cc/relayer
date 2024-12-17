package ika

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
)

// MockClient is a mock implementation of the Client interface for testing.
type MockClient struct{}

// NewMockClient creates a new mock Client instance.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// UpdateLC is a mock implementation used in tests.
func (p *MockClient) UpdateLC(_ context.Context, _ *tmtypes.LightBlock, _ zerolog.Logger) (models.SuiTransactionBlockResponse, error) {
	return models.SuiTransactionBlockResponse{}, nil
}

// ApproveAndSign is a mock implementation that returns a predefined signature.
func (p *MockClient) ApproveAndSign(_ context.Context, _, _ string, _ [][]byte) ([][]byte, error) {
	return [][]byte{{0x01, 0x02, 0x03}}, nil
}
