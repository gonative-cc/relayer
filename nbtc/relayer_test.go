package nbtc

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	err "github.com/gonative-cc/relayer/errors"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
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
	FetchFrom:             0,
}

func Test_Start(t *testing.T) {
	db := initTestDB(t)
	daltest.PopulateDB(t, db)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockFetcher, err := native2ika.NewMockAPISignRequestFetcher()
	assert.NilError(t, err)

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor, mockFetcher)
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
	mockFetcher, _ := native2ika.NewMockAPISignRequestFetcher()

	testCases := []struct {
		name            string
		db              *dal.DB
		nativeProcessor *native2ika.Processor
		btcProcessor    *ika2btc.Processor
		expectedError   error
		fetcher         native2ika.SignReqFetcher
	}{
		{
			name:            "DatabaseError",
			db:              nil,
			nativeProcessor: nativeProcessor,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoDB,
			fetcher:         mockFetcher,
		},
		{
			name:            "NativeProcessorError",
			db:              db,
			nativeProcessor: nil,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoNativeProcessor,
			fetcher:         mockFetcher,
		},
		{
			name:            "BtcProcessorError",
			db:              db,
			nativeProcessor: nativeProcessor,
			btcProcessor:    nil,
			expectedError:   err.ErrNoBtcProcessor,
			fetcher:         mockFetcher,
		},
		{
			name:            "BlockchainError",
			db:              db,
			nativeProcessor: nativeProcessor,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoFetcher,
			fetcher:         nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			relayer, err := NewRelayer(relayerConfig, tc.db, tc.nativeProcessor, tc.btcProcessor, tc.fetcher)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Assert(t, relayer == nil)
		})
	}
}

func TestRelayer_fetchAndStoreNativeSignRequests(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockFetcher, err := native2ika.NewMockAPISignRequestFetcher()
	assert.NilError(t, err)

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor, mockFetcher)
	assert.NilError(t, err)

	err = relayer.fetchAndStoreNativeSignRequests()
	assert.NilError(t, err)

	assert.Equal(t, relayer.signReqStart, 5) // Should be 5 after fetching 5 sign requests

	requests, err := db.GetPendingIkaSignRequests()
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 5) // Should be 5 inserted requests
}

func TestRelayer_storeSignRequest(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockFetcher, err := native2ika.NewMockAPISignRequestFetcher()
	assert.NilError(t, err)

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor, mockFetcher)
	assert.NilError(t, err)

	sr := native2ika.SignReq{ID: 1, Payload: []byte("rawTxBytes"), DWalletID: "dwallet1",
		UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()}

	err = relayer.storeSignRequest(sr)
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
