package bitcoinspv

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*Relayer, *clients.MockBTCClient, *clients.MockBitcoinSPV) {
	logger, _ := zap.NewDevelopment()
	btcClient := clients.NewMockBTCClient(t)
	lcClient := clients.NewMockBitcoinSPV(t)

	cfg := &config.RelayerConfig{
		// Add any necessary config values for testing
		Format:              "auto",
		Level:               "debug",
		NetParams:           "regtest",
		SleepDuration:       1 * time.Second,
		MaxSleepDuration:    10 * time.Second,
		BTCCacheSize:        1000,
		HeadersChunkSize:    10,
		ProcessBlockTimeout: 5 * time.Second,
	}

	relayer, err := New(
		cfg,
		logger,
		btcClient,
		lcClient,
		1*time.Second,  // retrySleepDuration
		10*time.Second, // maxRetrySleepDuration
		5*time.Second,  // processBlockTimeout
	)

	assert.NoError(t, err)
	assert.NotNil(t, relayer)

	return relayer, btcClient, lcClient
}

func TestNew(t *testing.T) {
	relayer, _, _ := setupTest(t)

	assert.NotNil(t, relayer.Config)
	assert.NotNil(t, relayer.logger)
	assert.NotNil(t, relayer.btcClient)
	assert.NotNil(t, relayer.lcClient)
	assert.Equal(t, time.Second, relayer.retrySleepDuration)
	assert.Equal(t, 10*time.Second, relayer.maxRetrySleepDuration)
	assert.Equal(t, 5*time.Second, relayer.processBlockTimeout)
	assert.NotNil(t, relayer.quitChannel)
	assert.False(t, relayer.isStarted)
}

func TestStartStop(t *testing.T) {
	relayer, btcClient, lcClient := setupTest(t)

	// Mock necessary method calls
	btcClient.On("SubscribeNewBlocks").Return()
	btcClient.On("BlockEventChannel").Maybe().Return(make(<-chan *btctypes.BlockEvent))
	btcClient.On("GetBTCTailBlocksByHeight", mock.Anything).Return([]*types.IndexedBlock{}, nil)

	firstBlockHash, _ := chainhash.NewHashFromStr("4a8cb347715524caa17d43987d527d0c11c0510f3e3c44a85035038a9b36e338")
	btcClient.On("GetBTCTipBlock").Return(firstBlockHash, int64(1), nil)

	firstBlockInfo := &clients.BlockInfo{
		Hash:   firstBlockHash,
		Height: int64(1),
	}
	lcClient.On("GetLatestBlockInfo", mock.Anything).Return(firstBlockInfo, nil)

	// Test Start
	relayer.Start()
	assert.True(t, relayer.isStarted)
	assert.False(t, relayer.isShutdown())

	// Test double start (should be no-op)
	relayer.Start()
	assert.True(t, relayer.isStarted)

	// Test Stop
	relayer.Stop()
	assert.True(t, relayer.isShutdown())

	// Test double stop (should be no-op)
	relayer.Stop()
	assert.True(t, relayer.isShutdown())

	// Cleanup
	btcClient.AssertExpectations(t)
	lcClient.AssertExpectations(t)
}

func TestRestartAfterShutdown(t *testing.T) {
	relayer, btcClient, lcClient := setupTest(t)

	// Mock necessary method calls
	btcClient.On("SubscribeNewBlocks").Return()
	btcClient.On("BlockEventChannel").Maybe().Return(make(<-chan *btctypes.BlockEvent))
	btcClient.On("GetBTCTailBlocksByHeight", mock.Anything).Return([]*types.IndexedBlock{}, nil)

	firstBlockHash, _ := chainhash.NewHashFromStr("4a8cb347715524caa17d43987d527d0c11c0510f3e3c44a85035038a9b36e338")
	btcClient.On("GetBTCTipBlock").Return(firstBlockHash, int64(1), nil)

	firstBlockInfo := &clients.BlockInfo{
		Hash:   firstBlockHash,
		Height: int64(1),
	}
	lcClient.On("GetLatestBlockInfo", mock.Anything).Return(firstBlockInfo, nil)

	// Start and stop the relayer
	relayer.Start()
	assert.True(t, relayer.isStarted)
	relayer.Stop()
	assert.True(t, relayer.isShutdown())

	// Start again after shutdown
	relayer.Start()
	assert.True(t, relayer.isStarted)

	// Cleanup
	btcClient.AssertExpectations(t)
	lcClient.AssertExpectations(t)
}
