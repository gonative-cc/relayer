package btcwrapper

import (
	"errors"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	Client
	mock.Mock
}

func (m *MockClient) GetBestBlockHash() (*chainhash.Hash, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chainhash.Hash), args.Error(1)
}

func (m *MockClient) GetBlockVerbose(hash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*btcjson.GetBlockVerboseResult), args.Error(1)
}

func (m *MockClient) GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wire.MsgBlock), args.Error(1)
}

func (m *MockClient) GetBlockHash(height int64) (*chainhash.Hash, error) {
	args := m.Called(height)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chainhash.Hash), args.Error(1)
}

func TestGetTipBlock(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.Client = Client{
		retrySleepDuration:    time.Millisecond,
		maxRetrySleepDuration: time.Millisecond * 10,
	}

	// Test successful case
	hash, _ := chainhash.NewHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
	expectedBlock := &btcjson.GetBlockVerboseResult{
		Hash:   hash.String(),
		Height: 0,
	}

	mockClient.On("GetBestBlockHash").Return(hash, nil).Once()
	mockClient.On("GetBlockVerbose", hash).Return(expectedBlock, nil).Once()

	block, err := mockClient.GetTipBlock()
	assert.NoError(t, err)
	assert.Equal(t, expectedBlock, block)

	// Test error in GetBestBlockHash
	mockClient.On("GetBestBlockHash").Return(nil, errors.New("network error")).Once()
	block, err = mockClient.GetTipBlock()
	assert.Error(t, err)
	assert.Nil(t, block)

	// Test error in GetBlockVerbose
	mockClient.On("GetBestBlockHash").Return(hash, nil).Once()
	mockClient.On("GetBlockVerbose", hash).Return(nil, errors.New("block error")).Once()
	block, err = mockClient.GetTipBlock()
	assert.Error(t, err)
	assert.Nil(t, block)
}
