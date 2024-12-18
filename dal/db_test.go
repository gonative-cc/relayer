package dal_test

import (
	"testing"
	"time"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"gotest.tools/v3/assert"
)

func TestInsertIkaSignRequest(t *testing.T) {
	db := daltest.InitTestDB(t)

	request := dal.IkaSignRequest{
		ID:        1,
		Payload:   []byte("payload"),
		DWalletID: "dwallet_id",
		UserSig:   "user_sig",
		FinalSig:  nil,
		Timestamp: uint64(time.Now().Unix()),
	}

	err := db.InsertIkaSignRequest(request)
	assert.NilError(t, err)

	retrievedReq, err := db.GetIkaSignRequest(1)
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedReq, &request)
}

// func TestGetSignedTxs(t *testing.T) {
// 	db := daltest.InitTestDB(t)
// 	daltest.PopulateDB(t, db)

// 	signedTxs, err := db.GetSignedTxs()
// 	assert.NilError(t, err)
// 	assert.Equal(t, len(signedTxs), 2)
// }

// func TestGetBroadcastedTxs(t *testing.T) {
// 	db := daltest.InitTestDB(t)
// 	daltest.PopulateDB(t, db)

// 	broadcastedTxs, err := db.GetBroadcastedTxs()
// 	assert.NilError(t, err)
// 	assert.Equal(t, len(broadcastedTxs), 2)
// }

// func TestUpdateTxStatus(t *testing.T) {
// 	db := daltest.InitTestDB(t)
// 	txID := uint64(1)
// 	tx := dal.Tx{
// 		BtcTxID: txID,
// 		Hash:    daltest.GetHashBytes(t, "1"),
// 		RawTx:   []byte("tx1-hex"),
// 		Status:  dal.StatusSigned,
// 	}
// 	err := db.InsertIkaSignRequest(tx)
// 	assert.NilError(t, err)

// 	err = db.UpdateTxStatus(txID, dal.StatusBroadcasted)
// 	assert.NilError(t, err)

// 	updatedTx, err := db.GetIkaSignRequest(txID)
// 	assert.NilError(t, err)
// 	assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
// }

// func TestInsertNativeTx(t *testing.T) {
// 	db := daltest.InitTestDB(t)

// 	tx := dal.NativeTx{
// 		TxID:           1,
// 		DWalletCapID:   "dwallet-cap-id",
// 		SignMessagesID: "sign-message-id",
// 		Messages:       [][]byte{[]byte("message1"), []byte("message2")},
// 		Status:         dal.Success,
// 	}

// 	err := db.InsertNativeTx(tx)
// 	assert.NilError(t, err)

// 	retrievedTxs, err := db.GetIkaTxsByStatus(dal.Success)
// 	assert.NilError(t, err)
// 	assert.Equal(t, len(retrievedTxs), 1)
// 	assert.Equal(t, retrievedTxs[0].TxID, tx.TxID)
// 	assert.Equal(t, retrievedTxs[0].DWalletCapID, tx.DWalletCapID)
// 	assert.Equal(t, retrievedTxs[0].SignMessagesID, tx.SignMessagesID)
// 	assert.Equal(t, len(retrievedTxs[0].Messages), 2)
// 	assert.Equal(t, retrievedTxs[0].Status, tx.Status)
// }

// func TestUpdateNativeTxStatus(t *testing.T) {
// 	db := daltest.InitTestDB(t)
// 	txID := uint64(1)

// 	tx := dal.NativeTx{
// 		TxID:           txID,
// 		DWalletCapID:   "dwallet-cap-id",
// 		SignMessagesID: "sign-message-id",
// 		Messages:       [][]byte{[]byte("message1")},
// 		Status:         dal.Success,
// 	}
// 	err := db.InsertNativeTx(tx)
// 	assert.NilError(t, err)

// 	err = db.UpdateNativeTxStatus(txID, dal.Failed)
// 	assert.NilError(t, err)

// 	updatedTx, err := db.GetIkaTx(txID)
// 	assert.NilError(t, err)
// 	assert.Equal(t, updatedTx.Status, dal.Failed)
// }

// func TestGetIkaTxsByStatus(t *testing.T) {
// 	db := daltest.InitTestDB(t)
// 	daltest.PopulateNativeDB(t, db)

// 	pendingTxs, err := db.GetIkaTxsByStatus(dal.Success)
// 	assert.NilError(t, err)
// 	assert.Equal(t, len(pendingTxs), 2)

// 	processedTxs, err := db.GetIkaTxsByStatus(dal.Failed)
// 	assert.NilError(t, err)
// 	assert.Equal(t, len(processedTxs), 1)
// }
