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
}

func Test_Start(t *testing.T) {
	db := initTestDB(t)
	daltest.PopulateDB(t, db)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, btcProcessor)
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

	testCases := []struct {
		name            string
		db              *dal.DB
		nativeProcessor *native2ika.Processor
		btcProcessor    *ika2btc.Processor
		expectedError   error
	}{
		{
			name:            "DatabaseError",
			db:              nil,
			nativeProcessor: nativeProcessor,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoDB,
		},
		{
			name:            "NativeProcessorError",
			db:              db,
			nativeProcessor: nil,
			btcProcessor:    btcProcessor,
			expectedError:   err.ErrNoNativeProcessor,
		},
		{
			name:            "BtcProcessorError",
			db:              db,
			nativeProcessor: nativeProcessor,
			btcProcessor:    nil,
			expectedError:   err.ErrNoBtcProcessor,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			relayer, err := NewRelayer(relayerConfig, tc.db, tc.nativeProcessor, tc.btcProcessor)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Assert(t, relayer == nil)
		})
	}
}

func initTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
