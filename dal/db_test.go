package dal_test

import (
	"testing"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"gotest.tools/v3/assert"
)

func TestInsertTx(t *testing.T) {
	db := daltest.InitTestDB(t)

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
	db := daltest.InitTestDB(t)

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

func TestGetBroadcastedTxs(t *testing.T) {
	db := daltest.InitTestDB(t)

	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: "tx1-hex", Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: "tx2-hex", Status: dal.StatusBroadcasted},
		{BtcTxID: 3, RawTx: "tx3-hex", Status: dal.StatusPending},
	}
	for _, tx := range transactions {
		err := db.InsertTx(tx)
		assert.NilError(t, err)
	}

	broadcastedTxs, err := db.GetBroadcastedTxs()
	assert.NilError(t, err)
	assert.Equal(t, len(broadcastedTxs), 1)
}

func TestUpdateTxStatus(t *testing.T) {
	db := daltest.InitTestDB(t)
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
