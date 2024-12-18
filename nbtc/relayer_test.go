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
	txs := daltest.PopulateDB(t, db)
	nativeTxs := daltest.PopulateNativeDB(t, db)

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

	for _, tx := range txs {

		updatedTx, err := db.GetTx(tx.BtcTxID)
		assert.NilError(t, err, "Error getting transaction")
		assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
	}

	time.Sleep(time.Second * 3)

	for _, tx := range txs {
		updatedTx, err := db.GetTx(tx.BtcTxID)
		if tx.BtcTxID == 1 {
			assert.NilError(t, err, "Error getting transaction")
			assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
		} else {
			assert.NilError(t, err, "Error getting transaction")
			assert.Equal(t, updatedTx.Status, dal.StatusConfirmed)
		}
	}

	for _, nativeTx := range nativeTxs {
		updatedTx, err := db.GetNativeTx(nativeTx.TxID)
		assert.NilError(t, err, "Error getting transaction")
		assert.Equal(t, updatedTx.Status, dal.NativeTxStatusProcessed)
	}

	relayer.Stop()
	relayer.db.Close()
}

func initTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
