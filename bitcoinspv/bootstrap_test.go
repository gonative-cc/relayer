package bitcoinspv

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupTestRelayer(t *testing.T) (*Relayer, *clients.MockBTCClient, *clients.MockBitcoinSPV) {
	btcClient := clients.NewMockBTCClient(t)
	lcClient := clients.NewMockBitcoinSPV(t)

	relayerConfig := &config.RelayerConfig{
		BTCCacheSize: 100,
	}

	logger := zap.NewNop().Sugar()

	r := &Relayer{
		btcClient:            btcClient,
		lcClient:             lcClient,
		Config:               relayerConfig,
		btcConfirmationDepth: 6,
		logger:               logger,
	}

	return r, btcClient, lcClient
}

func TestBootstrapRelayer(t *testing.T) {
	t.Run("successful bootstrap", func(t *testing.T) {
		r, btcClient, lcClient := setupTestRelayer(t)
		ctx := context.Background()

		// Mock GetBTCTipBlock
		btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(100), nil)

		// Mock GetLatestBlockInfo
		lcClient.On("GetLatestBlockInfo", mock.Anything).Return(&clients.BlockInfo{
			Height: 90,
		}, nil)

		// Mock GetBTCTailBlocksByHeight
		blocks := make([]*types.IndexedBlock, 6)
		for i := int64(85); i <= 90; i++ {
			blocks[i-85] = &types.IndexedBlock{
				BlockHeader: &wire.BlockHeader{},
				BlockHeight: i,
			}
		}
		btcClient.On("GetBTCTailBlocksByHeight", int64(85)).Return(blocks, nil)
		lcClient.On("ContainsBlock", mock.Anything, mock.Anything).Return(true, nil)

		// Mock SubscribeNewBlocks
		btcClient.On("SubscribeNewBlocks").Return()

		err := r.bootstrapRelayer(ctx, false)
		assert.NoError(t, err)
	})

	t.Run("failed BTC sync", func(t *testing.T) {
		r, btcClient, _ := setupTestRelayer(t)
		ctx := context.Background()

		// Mock GetBTCTipBlock with error
		expectedErr := errors.New("btc sync failed")
		btcClient.On("GetBTCTipBlock").Return(nil, int64(0), expectedErr)

		err := r.bootstrapRelayer(ctx, false)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("failed light client sync", func(t *testing.T) {
		r, btcClient, lcClient := setupTestRelayer(t)
		ctx := context.Background()

		// Mock GetBTCTipBlock
		btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(100), nil)

		// Mock GetLatestBlockInfo with error
		expectedErr := errors.New("light client sync failed")
		lcClient.On("GetLatestBlockInfo", mock.Anything).Return(nil, expectedErr)

		err := r.bootstrapRelayer(ctx, false)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestWaitForBTCCatchup(t *testing.T) {
	t.Run("successful catchup", func(t *testing.T) {
		r, btcClient, lcClient := setupTestRelayer(t)
		ctx := context.Background()

		// Set shorter ticker for test
		originalTicker := bootstrapSyncTicker
		bootstrapSyncTicker = 100 * time.Millisecond
		defer func() { bootstrapSyncTicker = originalTicker }()

		// First call: BTC behind
		firstCall := btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(90), nil).Once()
		// Second call: BTC caught up
		btcClient.On("GetBTCTipBlock").Return(&chainhash.Hash{}, int64(100), nil).Once().NotBefore(firstCall)

		// LC height stays constant
		lcClient.On("GetLatestBlockInfo", mock.Anything).Return(&clients.BlockInfo{
			Height: 95,
		}, nil)

		err := r.waitForBTCCatchup(ctx, 90, 95)
		assert.NoError(t, err)
	})
}

func TestInitializeBTCCache(t *testing.T) {
	t.Run("successful cache initialization", func(t *testing.T) {
		r, btcClient, lcClient := setupTestRelayer(t)
		ctx := context.Background()

		lcClient.On("GetLatestBlockInfo", mock.Anything).Return(&clients.BlockInfo{
			Height: 100,
		}, nil)

		blocks := make([]*types.IndexedBlock, 6)
		for i := int64(95); i <= 100; i++ {
			blocks[i-95] = &types.IndexedBlock{
				BlockHeader: &wire.BlockHeader{},
				BlockHeight: i,
			}
		}
		btcClient.On("GetBTCTailBlocksByHeight", int64(95)).Return(blocks, nil)

		err := r.initializeBTCCache(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, r.btcCache)
	})

	t.Run("failed cache initialization - invalid size", func(t *testing.T) {
		r, _, _ := setupTestRelayer(t)
		ctx := context.Background()

		r.Config.BTCCacheSize = 0
		err := r.initializeBTCCache(ctx)
		assert.Error(t, err)
	})

	t.Run("failed cache initialization - GetBTCTailBlocksByHeight error", func(t *testing.T) {
		r, btcClient, lcClient := setupTestRelayer(t)
		ctx := context.Background()

		lcClient.On("GetLatestBlockInfo", mock.Anything).Return(&clients.BlockInfo{
			Height: 100,
		}, nil)

		expectedErr := errors.New("failed to get tail blocks")
		btcClient.On("GetBTCTailBlocksByHeight", int64(95)).Return(nil, expectedErr)

		err := r.initializeBTCCache(ctx)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
