package daltest

import (
	"testing"

	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/v3/assert"
)

func TestInsertTx(t *testing.T) {
	db := InitTestDB(t)

	tx := dal.Tx{
		BtcTxID: 1,
		RawTx:   "raw-transaction-hex",
		Status:  dal.StatusPending,
	}

	err := db.InsertTx(tx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetTx(1)
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &tx)
}

func TestGetPendingTxs(t *testing.T) {
	db := InitTestDB(t)

	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: "tx1-hex", Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: "tx2-hex", Status: dal.StatusBroadcasted},
		{BtcTxID: 3, RawTx: "tx3-hex", Status: dal.StatusPending},
	}
	for _, tx := range transactions {
		err := db.InsertTx(tx)
		assert.NilError(t, err)
	}

	pendingTxs, err := db.GetPendingTxs()
	assert.NilError(t, err)
	assert.Equal(t, len(pendingTxs), 2)
}

func TestUpdateTxStatus(t *testing.T) {
	db := InitTestDB(t)
	txID := uint64(1)
	tx := dal.Tx{
		BtcTxID: txID,
		RawTx:   "raw-transaction-hex",
		Status:  dal.StatusPending,
	}
	err := db.InsertTx(tx)
	assert.NilError(t, err)

	err = db.UpdateTxStatus(txID, dal.StatusBroadcasted)
	assert.NilError(t, err)

	updatedTx, err := db.GetTx(txID)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
}

// InitTestDB initializes an in-memory database for testing purposes.
func InitTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
