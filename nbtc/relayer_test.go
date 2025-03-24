package nbtc

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"gotest.tools/v3/assert"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/ika2btc/bitcoin"
	"github.com/gonative-cc/relayer/native"
)

// testSuite holds the common dependencies
type testSuite struct {
	db             dal.DB
	ikaClient      *ika.MockClient
	btcProcessor   *ika2btc.Processor
	signReqFetcher *native.APISignRequestFetcher
	relayer        *Relayer
	ctx            context.Context
	cancel         context.CancelFunc
}

var btcClientConfig = rpcclient.ConnConfig{
	Host:         "test_rpc",
	User:         "test_user",
	Pass:         "test_pass",
	HTTPPostMode: true,
	DisableTLS:   false,
}

var relayerConfig = RelayerConfig{
	ProcessTxsInterval: time.Second * 5,
	ConfirmTxsInterval: time.Second * 7,
	SignReqFetchFrom:   0,
	SignReqFetchLimit:  5,
}

// setupTestProcessor initializes the common dependencies
func setupTestSuite(t *testing.T) *testSuite {
	ctx := context.Background()
	db := daltest.InitTestDB(ctx, t)
	ikaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	signReqFetcher, err := native.NewMockAPISignRequestFetcher()
	assert.NilError(t, err)

	relayer, err := NewRelayer(relayerConfig, db, btcProcessor, signReqFetcher)
	assert.NilError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	return &testSuite{
		db:             db,
		ikaClient:      ikaClient,
		btcProcessor:   btcProcessor,
		signReqFetcher: signReqFetcher,
		relayer:        relayer,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func Test_Start(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.cancel()

	daltest.PopulateDB(ts.ctx, t, ts.db)

	// Start the relayer in a separate goroutine
	go func() {
		assert.NilError(t, ts.relayer.Start(ts.ctx))
	}()

	t.Run("Transaction Broadcasted", func(t *testing.T) {
		time.Sleep(time.Second * 6)
		confirmedTx, err := ts.db.GetBitcoinTx(ts.ctx, 2, daltest.DecodeBTCHash(t, "0"))
		assert.NilError(t, err)
		assert.Equal(t, confirmedTx.Status, dal.Broadcasted)
	})

	t.Run("Transaction Confirmed", func(t *testing.T) {
		time.Sleep(time.Second * 3) // Give time for confirmation
		confirmedTx, err := ts.db.GetBitcoinTx(ts.ctx, 2, daltest.DecodeBTCHash(t, "0"))
		assert.NilError(t, err)
		assert.Equal(t, confirmedTx.Status, dal.Confirmed)
	})
	ts.db.Close()
}

func TestNewRelayer_ErrorCases(t *testing.T) {
	ts := setupTestSuite(t)
	testCases := []struct {
		db            dal.DB
		expectedError error
		fetcher       native.SignReqFetcher
		btcProcessor  *ika2btc.Processor
		name          string
	}{
		{
			name:          "NativeProcessorError",
			db:            ts.db,
			btcProcessor:  ts.btcProcessor,
			expectedError: native.ErrNoNativeProcessor,
			fetcher:       ts.signReqFetcher,
		}, {
			name:          "BtcProcessorError",
			db:            ts.db,
			btcProcessor:  nil,
			expectedError: bitcoin.ErrNoBtcProcessor,
			fetcher:       ts.signReqFetcher,
		}, {
			name:          "BlockchainError",
			db:            ts.db,
			btcProcessor:  ts.btcProcessor,
			expectedError: native.ErrNoFetcher,
			fetcher:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			relayer, err := NewRelayer(relayerConfig, tc.db, tc.btcProcessor, tc.fetcher)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Assert(t, relayer == nil)
		})
	}
	ts.db.Close()
}

func TestRelayer_fetchAndStoreNativeSignRequests(t *testing.T) {
	ts := setupTestSuite(t)
	ctx := context.Background()

	err := ts.relayer.fetchAndStoreNativeSignRequests(ctx)
	assert.NilError(t, err)
	assert.Equal(t, ts.relayer.signReqFetchFrom, 5) // Should be 5 after fetching 5 sign requests

	requests, err := ts.db.GetPendingIkaSignRequests(ts.ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 5) // Should be 5 inserted requests
	ts.db.Close()
}

func TestRelayer_storeSignRequest(t *testing.T) {
	ts := setupTestSuite(t)
	ctx := context.Background()

	requests, err := ts.db.GetPendingIkaSignRequests(ts.ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 0)

	sr := native.SignReq{ID: 1, Payload: []byte("rawTxBytes"), DWalletID: "dwallet1",
		UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()}

	err = ts.relayer.storeSignRequest(ctx, sr)
	assert.NilError(t, err)

	requests, err = ts.db.GetPendingIkaSignRequests(ts.ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 1)
	ts.db.Close()
}
