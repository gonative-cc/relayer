package nbtc

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/assert"
)

var config = rpcclient.ConnConfig{
	Host:         "test_rpc",
	User:         "test_user",
	Pass:         "test_pass",
	HTTPPostMode: true,
	DisableTLS:   false,
}

var rawTxBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xf2, 0x05,
	0x2a, 0x01, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x88, 0xac, 0x00, 0x00,
	0x00, 0x00,
}

func Test_Start(t *testing.T) {
	db := initTestDB(t)

	relayer, err := NewRelayer(config, 0, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: rawTxBytes, Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: rawTxBytes, Status: dal.StatusPending},
	}
	for _, tx := range transactions {
		err = db.InsertTx(tx)
		if err != nil {
			assert.NilError(t, err)
		}
	}

	// Start the relayer in a separate goroutine
	go func() {
		err := relayer.Start()
		assert.NilError(t, err)
	}()

	time.Sleep(time.Second * 3)


	time.Sleep(time.Second * 6)

	relayer.Stop()

	for _, tx := range transactions {

		updatedTx, err := db.GetTx(tx.BtcTxID)
		assert.NilError(t, err, "Error getting transaction")
		assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
	}

	time.Sleep(time.Second * 5)

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
	relayer, err := NewRelayer(db)

	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: rawTxBytes, Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: rawTxBytes, Status: dal.StatusPending},
	}
	for _, tx := range transactions {
		err := db.InsertTx(tx)
		assert.NilError(t, err)
	}

	relayer, err := NewRelayer(config, 0, db)
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
	relayer, err := NewRelayer(db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	relayer.checkConfirmations()

	time.Sleep(time.Millisecond * 500)

	updatedTx1, err := db.GetTx(1)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx1.Status, dal.StatusBroadcasted)

	updatedTx2, err := db.GetTx(2)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx2.Status, dal.StatusConfirmed)

}
  
func Test_NewRelayer_DatabaseError(t *testing.T) {
	relayer, err := NewRelayer(config, 0, nil)
	assert.ErrorContains(t, err, "database cannot be nil")
	assert.Assert(t, relayer == nil)
}

func Test_NewRelayer_MissingEnvVatiables(t *testing.T) {
	db := initTestDB(t)
	config.Host = ""
	relayer, err := NewRelayer(config, 0, db)
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
