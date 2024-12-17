package nbtc

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
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

// TODO: update this test
func Test_Start(t *testing.T) {
	db := initTestDB(t)
	txs := daltest.PopulateDB(t, db)

	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	// Start the relayer in a separate goroutine
	go func() {
		err := relayer.Start()
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
