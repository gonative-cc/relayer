package nbtc

import (
	"os"
	"testing"
	"time"

	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/assert"
)

func TestRelayer_Start(t *testing.T) {
	db := initTestDB(t)

	relayer, err := NewRelayer(db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0200f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: dal.StatusPending},
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

	relayer.Stop()

	for _, tx := range transactions {
		updatedTx, err := db.GetTx(tx.BtcTxID)
		assert.NilError(t, err, "Error getting transaction")
		assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
	}

	relayer.db.Close()
}

func TestRelayer_processPendingTxs(t *testing.T) {
	db := initTestDB(t)

	// Insert mock transactions with "pending" status
	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0200f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: dal.StatusPending},
	}
	for _, tx := range transactions {
		err := db.InsertTx(tx)
		assert.NilError(t, err)
	}

	relayer, err := NewRelayer(db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	err = relayer.processPendingTxs()
	assert.NilError(t, err)

	for _, tx := range transactions {
		updatedTx, err := db.GetTx(tx.BtcTxID)
		assert.NilError(t, err)
		assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
	}
}

func TestNewRelayer_DatabaseError(t *testing.T) {
	relayer, err := NewRelayer(nil)
	assert.ErrorContains(t, err, "database cannot be nil")
	assert.Assert(t, relayer == nil)
}

func TestNewRelayer_MissingEnvVatiables(t *testing.T) {
	// Clear the env variables
	os.Unsetenv("BTC_RPC")
	os.Unsetenv("BTC_RPC_USER")
	os.Unsetenv("BTC_RPC_PASS")
	db := initTestDB(t)
	relayer, err := NewRelayer(db)
	assert.ErrorContains(t, err, "missing env variables with Bitcoin node configuration")
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
