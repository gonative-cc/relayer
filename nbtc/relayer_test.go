package nbtc

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native2ika"
	"gotest.tools/assert"
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

func Test_NewRelayer_DatabaseError(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)

	relayer, err := NewRelayer(relayerConfig, nil, nativeProcessor, btcProcessor)
	assert.ErrorContains(t, err, "database cannot be nil")
	assert.Assert(t, relayer == nil)
}

func Test_NewRelayer_NativeProcessorError(t *testing.T) {
	db := initTestDB(t)
	btcProcessor, _ := ika2btc.NewProcessor(btcClientConfig, 6, db)
	btcProcessor.BtcClient = &bitcoin.MockClient{}

	relayer, err := NewRelayer(relayerConfig, db, nil, btcProcessor)
	assert.ErrorContains(t, err, "nativeProcessor cannot be nil")
	assert.Assert(t, relayer == nil)
}

func Test_NewRelayer_BtcProcessorError(t *testing.T) {
	db := initTestDB(t)
	mockIkaClient := ika.NewMockClient()
	nativeProcessor := native2ika.NewProcessor(mockIkaClient, db)

	relayer, err := NewRelayer(relayerConfig, db, nativeProcessor, nil)
	assert.ErrorContains(t, err, "btcProcessor cannot be nil")
	assert.Assert(t, relayer == nil)
}

func initTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
