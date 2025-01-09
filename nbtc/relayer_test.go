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

// testSuite holds the common dependencies
type testSuite struct {
	db              *dal.DB
	mockIkaClient   *ika.MockClient
	btcProcessor    *ika2btc.Processor
	nativeProcessor *native2ika.Processor
	mockFetcher     *native2ika.APISignRequestFetcher
}

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
	FetchLimit:            5,
}

// setupTestProcessor initializes the common dependencies
func setupTestSuite(t *testing.T) *testSuite {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)
	mockFetcher, err := native2ika.NewMockAPISignRequestFetcher()
	if err != nil {
		t.Fatal(err)
	}

	return &testSuite{
		db:              db,
		mockIkaClient:   mockIkaClient,
		btcProcessor:    btcProcessor,
		nativeProcessor: nativeProcessor,
		mockFetcher:     mockFetcher,
	}
}

func Test_Start(t *testing.T) {
	ts := setupTestSuite(t)
	daltest.PopulateDB(t, ts.db)

	relayer, err := NewRelayer(relayerConfig, ts.db, ts.nativeProcessor, ts.btcProcessor, ts.mockFetcher)
	assert.NilError(t, err)

	ctx := context.Background()

	// Start the relayer in a separate goroutine
	go func() {
		err := relayer.Start(ctx)
		assert.NilError(t, err)
	}()

	time.Sleep(time.Second * 6)

	confirmedTx, err := ts.db.GetBitcoinTx(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Broadcasted)

	time.Sleep(time.Second * 3)

	confirmedTx, err = ts.db.GetBitcoinTx(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Confirmed)

	relayer.Stop()
	relayer.db.Close()
}

func TestNewRelayer_ErrorCases(t *testing.T) {
	ts := setupTestSuite(t)
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
			nativeProcessor: ts.nativeProcessor,
			btcProcessor:    ts.btcProcessor,
			expectedError:   err.ErrNoDB,
			fetcher:         ts.mockFetcher,
		},
		{
			name:            "NativeProcessorError",
			db:              ts.db,
			nativeProcessor: nil,
			btcProcessor:    ts.btcProcessor,
			expectedError:   err.ErrNoNativeProcessor,
			fetcher:         ts.mockFetcher,
		},
		{
			name:            "BtcProcessorError",
			db:              ts.db,
			nativeProcessor: ts.nativeProcessor,
			btcProcessor:    nil,
			expectedError:   err.ErrNoBtcProcessor,
			fetcher:         ts.mockFetcher,
		},
		{
			name:            "BlockchainError",
			db:              ts.db,
			nativeProcessor: ts.nativeProcessor,
			btcProcessor:    ts.btcProcessor,
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
	ts := setupTestSuite(t)
	relayer, err := NewRelayer(relayerConfig, ts.db, ts.nativeProcessor, ts.btcProcessor, ts.mockFetcher)
	assert.NilError(t, err)

	err = relayer.fetchAndStoreNativeSignRequests()
	assert.NilError(t, err)

	assert.Equal(t, relayer.signReqStart, 5) // Should be 5 after fetching 5 sign requests

	requests, err := ts.db.GetPendingIkaSignRequests()
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 5) // Should be 5 inserted requests
}

func TestRelayer_storeSignRequest(t *testing.T) {
	ts := setupTestSuite(t)
	relayer, err := NewRelayer(relayerConfig, ts.db, ts.nativeProcessor, ts.btcProcessor, ts.mockFetcher)
	assert.NilError(t, err)

	sr := native2ika.SignReq{ID: 1, Payload: []byte("rawTxBytes"), DWalletID: "dwallet1",
		UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()}

	err = relayer.storeSignRequest(sr)
	assert.NilError(t, err)

	requests, err := ts.db.GetPendingIkaSignRequests()
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
