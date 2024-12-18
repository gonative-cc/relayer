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

func Test_processPendingTxs(t *testing.T) {
	db := initTestDB(t)
	txs := daltest.PopulateDB(t, db)
	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	err = relayer.processPendingTxs()
	assert.NilError(t, err)

	for _, tx := range txs {
		updatedTx, err := db.GetTx(tx.BtcTxID)
		assert.NilError(t, err)
		assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
	}
}

func Test_checkConfirmations(t *testing.T) {
	db := initTestDB(t)
	daltest.PopulateDB(t, db)
	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	relayer.checkConfirmations()

	updatedTx1, err := db.GetTx(1)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx1.Status, dal.StatusBroadcasted)

	updatedTx2, err := db.GetTx(2)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx2.Status, dal.StatusConfirmed)

}

func Test_NewRelayer_DatabaseError(t *testing.T) {
	relayer, err := NewRelayer(btcClientConfig, relayerConfig, nil)
	assert.ErrorContains(t, err, "database cannot be nil")
	assert.Assert(t, relayer == nil)
}

func Test_NewRelayer_MissingEnvVatiables(t *testing.T) {
	db := initTestDB(t)
	btcClientConfig.Host = ""
	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.ErrorContains(t, err, "missing bitcoin node configuration")
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
