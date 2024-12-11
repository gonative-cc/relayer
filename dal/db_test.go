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
		Status:  dal.StatusSigned,
		Hash:    daltest.GetHashBytes(t, "1"),
		RawTx:   []byte("tx1-hex"),
	}

	err := db.InsertTx(tx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetTx(1)
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &tx)
}

func TestGetSignedTxs(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateDB(t, db)

	signedTxs, err := db.GetSignedTxs()
	assert.NilError(t, err)
	assert.Equal(t, len(signedTxs), 2)
}

func TestGetBroadcastedTxs(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateDB(t, db)

	broadcastedTxs, err := db.GetBroadcastedTxs()
	assert.NilError(t, err)
	assert.Equal(t, len(broadcastedTxs), 2)
}

func TestUpdateTxStatus(t *testing.T) {
	db := daltest.InitTestDB(t)
	txID := uint64(1)
	tx := dal.Tx{
		BtcTxID: txID,
		Hash:    daltest.GetHashBytes(t, "1"),
		RawTx:   []byte("tx1-hex"),
		Status:  dal.StatusSigned,
	}
	err := db.InsertTx(tx)
	assert.NilError(t, err)

	err = db.UpdateTxStatus(txID, dal.StatusBroadcasted)
	assert.NilError(t, err)

	updatedTx, err := db.GetTx(txID)
	assert.NilError(t, err)
	assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
}
