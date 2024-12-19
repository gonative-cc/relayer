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
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)
	daltest.PopulateIkaTxs(t, db)

	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	// Start the relayer in a separate goroutine
	go func() {
		err := relayer.Start()
		assert.NilError(t, err)
	}()

	time.Sleep(time.Second * 6)

	confirmedTx, err := db.GetBitcoinTxByTxIDAndBtcTxID(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Broadcasted)

	time.Sleep(time.Second * 3)

	confirmedTx, err = db.GetBitcoinTxByTxIDAndBtcTxID(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Confirmed)

	relayer.Stop()
	relayer.db.Close()
}

func Test_processSignedTxs(t *testing.T) {
	db := initTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)
	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	err = relayer.processSignedTxs()
	assert.NilError(t, err)

	updatedTx, err := db.GetBitcoinTxByTxIDAndBtcTxID(2, daltest.GetHashBytes(t, "0"))
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.Broadcasted)
}

func Test_checkConfirmations(t *testing.T) {
	db := initTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)
	relayer, err := NewRelayer(btcClientConfig, relayerConfig, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	relayer.checkConfirmations()

	uupdatedTx, err := db.GetBitcoinTxByTxIDAndBtcTxID(4, daltest.GetHashBytes(t, "3"))
	assert.NilError(t, err)
	assert.Equal(t, uupdatedTx.Status, dal.Confirmed)

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
