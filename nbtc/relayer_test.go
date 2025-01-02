package nbtc

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	err "github.com/gonative-cc/relayer/errors"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native"
	"github.com/gonative-cc/relayer/native2ika"
	"gotest.tools/v3/assert"
)

var btcClientConfig = rpcclient.ConnConfig{
	Host:         "test_rpc",
	User:         "test_user",
	Pass:         "test_pass",
	HTTPPostMode: true,
	DisableTLS:   false,
}

var relayerConfig = RelayerConfig{
	ProcessTxsInterval:    time.Second * 5,
	ConfirmTxsInterval:    time.Second * 7,
	ConfirmationThreshold: 6,
	BlockHeigh:            1,
}

func Test_Start(t *testing.T) {
	db := initTestDB(t)
	daltest.PopulateDB(t, db)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockBlockchain := native.MockBlockchain{}

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor, mockBlockchain)
	assert.NilError(t, err)

	ctx := context.Background()

	// Start the relayer in a separate goroutine
	go func() {
		err := relayer.Start(ctx)
		assert.NilError(t, err)
	}()

	time.Sleep(time.Second * 6)

	confirmedTx, err := db.GetBitcoinTx(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Broadcasted)

	time.Sleep(time.Second * 3)

	confirmedTx, err = db.GetBitcoinTx(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Confirmed)

	relayer.Stop()
	relayer.db.Close()
}

func TestNewRelayer_ErrorCases(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockBlockchain := native.MockBlockchain{}

	testCases := []struct {
		name            string
		db              *dal.DB
		nativeProcessor *native2ika.Processor
		btcProcessor    *ika2btc.Processor
		expectedError   error
		blockchain      native.Blockchain
	}{
		{
			name:            "DatabaseError",
			db:              nil,
			nativeProcessor: nativeProcessor,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoDB,
			blockchain:      mockBlockchain,
		},
		{
			name:            "NativeProcessorError",
			db:              db,
			nativeProcessor: nil,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoNativeProcessor,
			blockchain:      mockBlockchain,
		},
		{
			name:            "BtcProcessorError",
			db:              db,
			nativeProcessor: nativeProcessor,
			btcProcessor:    nil,
			expectedError:   err.ErrNoBtcProcessor,
			blockchain:      mockBlockchain,
		},
		{
			name:            "BlockchainError",
			db:              db,
			nativeProcessor: nativeProcessor,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoBlockchain,
			blockchain:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			relayer, err := NewRelayer(relayerConfig, tc.db, tc.nativeProcessor, tc.btcProcessor, tc.blockchain)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Assert(t, relayer == nil)
		})
	}
}

func TestRelayer_fetchAndProcessNativeBlocks(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockBlockchain := native.MockBlockchain{}

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor, mockBlockchain)
	assert.NilError(t, err)

	err = relayer.fetchAndProcessNativeBlocks(context.Background())
	assert.NilError(t, err)

	assert.Equal(t, relayer.fetchedBlockHeight, int64(20)) // Should be 20 after fetching 20 blocks

	requests, err := db.GetPendingIkaSignRequests()
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 20) // Should be 20 inserted requests
}

func TestRelayer_processNativeBlock(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockBlockchain := native.MockBlockchain{}

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor, mockBlockchain)
	assert.NilError(t, err)

	block := &tmtypes.Block{
		Header: tmtypes.Header{Height: 1},
	}

	err = relayer.processNativeBlock(block)
	assert.NilError(t, err)

	requests, err := db.GetPendingIkaSignRequests()
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 1)
}

func initTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
