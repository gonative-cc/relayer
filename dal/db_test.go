package dal

import (
	"testing"

	"gotest.tools/assert"
)

func TestInsertTx(t *testing.T) {
	db := initTestDB(t)

	tx := Tx{
		BtcTxID: 1,
		RawTx:   "raw-transaction-hex",
		Status:  StatusPending,
	}

	err := db.InsertTx(tx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetTx(1)
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &tx)
}

func TestGetPendingTxs(t *testing.T) {
	db := initTestDB(t)

	transactions := []Tx{
		{BtcTxID: 1, RawTx: "tx1-hex", Status: StatusPending},
		{BtcTxID: 2, RawTx: "tx2-hex", Status: StatusBroadcasted},
		{BtcTxID: 3, RawTx: "tx3-hex", Status: StatusPending},
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
	db := initTestDB(t)

	txID := uint64(1)

	tx := Tx{
		BtcTxID: txID,
		RawTx:   "raw-transaction-hex",
		Status:  StatusPending,
	}
	err := db.InsertTx(tx)
	assert.NilError(t, err)

	err = db.UpdateTxStatus(txID, StatusBroadcasted)
	assert.NilError(t, err)

	updatedTx, err := db.GetTx(txID)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, StatusBroadcasted)
}

func initTestDB(t *testing.T) *DB {
	t.Helper()

	db, err := NewDB(":memory:")
	assert.NilError(t, err)
	return db
}
