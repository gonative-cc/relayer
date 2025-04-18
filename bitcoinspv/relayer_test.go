package bitcoinspv

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gonative-cc/relayer/bitcoinspv/clients"
	"github.com/gonative-cc/relayer/bitcoinspv/clients/mocks"
	"github.com/gonative-cc/relayer/bitcoinspv/config"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTest(t *testing.T) (*Relayer, *mocks.MockBTCClient, *mocks.MockBitcoinSPV) {
	t.Helper()
        var (
	     logger = zerolog.Nop()
	     btcClient = mocks.NewMockBTCClient(t)
	     lcClient = mocks.NewMockBitcoinSPV(t)
	)

	cfg := &config.RelayerConfig{
		Format:                "auto",
		Level:                 "debug",
		NetParams:             "regtest",
		RetrySleepDuration:    1 * time.Second,
		MaxRetrySleepDuration: 10 * time.Second,
		BTCCacheSize:          1000,
		HeadersChunkSize:      10,
		ProcessBlockTimeout:   5 * time.Second,
	}

	relayer, err := New(
		cfg,
		logger,
		btcClient,
		lcClient,
	)

	assert.NoError(t, err)
	assert.NotNil(t, relayer)

	return relayer, btcClient, lcClient
}

func setupMocks(t *testing.T, btcClient *mocks.MockBTCClient, lcClient *mocks.MockBitcoinSPV) {
	t.Helper()
	firstBlockHash, _ := chainhash.NewHashFromStr("4a8cb347715524caa17d43987d527d0c11c0510f3e3c44a85035038a9b36e338")
	firstBlockInfo := &clients.BlockInfo{
		Hash:   firstBlockHash,
		Height: int64(1),
	}

	btcClient.On("GetBTCTipBlock").Return(firstBlockHash, int64(1), nil)
	lcClient.On("GetLatestBlockInfo", mock.Anything).Return(firstBlockInfo, nil)
	btcClient.On("GetBTCTailBlocksByHeight", mock.Anything, mock.Anything).Return([]*types.IndexedBlock{}, nil)
	btcClient.On("SubscribeNewBlocks").Return()
	btcClient.On("BlockEventChannel").Maybe().Return(make(<-chan *btctypes.BlockEvent))
}

func cleanupRelayer(t *testing.T, relayer *Relayer) {
	t.Helper()
	assert.NotPanics(t, func() { relayer.Stop() })
	assert.True(t, relayer.isShutdown())
	assert.False(t, relayer.isRunning())
	waitDone := make(chan struct{})
	go func() {
		relayer.WaitForShutdown()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for WaitForShutdown")
	}
}

func TestNew(t *testing.T) {
	relayer, _, _ := setupTest(t)

	assert.NotNil(t, relayer.Config)
	assert.NotNil(t, relayer.logger)
	assert.NotNil(t, relayer.btcClient)
	assert.NotNil(t, relayer.lcClient)
	assert.Equal(t, time.Second, relayer.Config.RetrySleepDuration)
	assert.Equal(t, 10*time.Second, relayer.Config.MaxRetrySleepDuration)
	assert.Equal(t, 5*time.Second, relayer.Config.ProcessBlockTimeout)
	assert.NotNil(t, relayer.quitChannel)
	assert.False(t, relayer.isStarted)
}

func TestIsRunning(t *testing.T) {
	relayer, _, _ := setupTest(t)
	assert.False(t, relayer.isRunning())

	relayer.isStarted = true
	assert.True(t, relayer.isRunning())

	relayer.isStarted = false
	assert.False(t, relayer.isRunning())
}

func TestIsShutdown(t *testing.T) {
	relayer, _, _ := setupTest(t)
	assert.False(t, relayer.isShutdown())

	// simulate shutdown
	close(relayer.quitChannel)
	assert.True(t, relayer.isShutdown())

	// simulate restart
	relayer.quitChannel = make(chan struct{})
	assert.False(t, relayer.isShutdown())
}

func TestStop(t *testing.T) {
	relayer, _, _ := setupTest(t)
	initialChan := relayer.quitChan()

	// initial state
	assert.False(t, relayer.isShutdown())
	relayer.isStarted = true
	assert.True(t, relayer.isRunning())

	relayer.Stop()

	// after calling stop
	assert.True(t, relayer.isShutdown())
	assert.False(t, relayer.isRunning())
	_, chanOpen := <-initialChan
	assert.False(t, chanOpen)

	// call stop again
	assert.NotPanics(t, func() { relayer.Stop() })
	assert.True(t, relayer.isShutdown())
	assert.False(t, relayer.isRunning())
}

func TestStart(t *testing.T) {
	relayer, btcClient, lcClient := setupTest(t)
	setupMocks(t, btcClient, lcClient)
	t.Cleanup(func() {
		cleanupRelayer(t, relayer)
	})
	assert.False(t, relayer.isRunning())

	go relayer.Start()
	time.Sleep(10 * time.Millisecond)

	assert.False(t, relayer.isShutdown())
	assert.True(t, relayer.isRunning())

	// call start again
	assert.NotPanics(t, func() { relayer.Start() })

	assert.False(t, relayer.isShutdown())
	assert.True(t, relayer.isRunning())

	btcClient.AssertExpectations(t)
	lcClient.AssertExpectations(t)
}

func TestRestartAfterShutdown(t *testing.T) {
	relayer, btcClient, lcClient := setupTest(t)
	setupMocks(t, btcClient, lcClient)
	t.Cleanup(func() {
		cleanupRelayer(t, relayer)
	})
	assert.False(t, relayer.isRunning())

	go relayer.Start()
	time.Sleep(10 * time.Millisecond)

	assert.True(t, relayer.isRunning())
	assert.False(t, relayer.isShutdown())

	assert.NotPanics(t, func() { relayer.Stop() })
	assert.True(t, relayer.isShutdown())
	assert.False(t, relayer.isRunning())

	// Start again after shutdown
	go relayer.Start()
	time.Sleep(10 * time.Millisecond)

	assert.True(t, relayer.isRunning())
	assert.False(t, relayer.isShutdown())

	btcClient.AssertExpectations(t)
	lcClient.AssertExpectations(t)
}
