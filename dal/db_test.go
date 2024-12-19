package dal_test

import (
	"testing"
	"time"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"gotest.tools/v3/assert"
)

func Test_InsertIkaSignRequest(t *testing.T) {
	db := daltest.InitTestDB(t)

	request := dal.IkaSignRequest{
		ID:        1,
		Payload:   []byte("payload"),
		DWalletID: "dwallet_id",
		UserSig:   "user_sig",
		FinalSig:  nil,
		Timestamp: time.Now().Unix(),
	}

	err := db.InsertIkaSignRequest(request)
	assert.NilError(t, err)

	retrievedReq, err := db.GetIkaSignRequestByID(1)
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedReq, &request)
}

func Test_InsertIkaTx(t *testing.T) {
	db := daltest.InitTestDB(t)

	ikaTx := dal.IkaTx{
		TxID: 1, Status: dal.Success, IkaTxID: "ika_tx_1", Timestamp: time.Now().Unix(), Note: "",
	}

	err := db.InsertIkaTx(ikaTx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetIkaTxByTxIDAndIkaTxID(1, "ika_tx_1")
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &ikaTx)
}

func Test_InsertBitcoinTx(t *testing.T) {
	db := daltest.InitTestDB(t)

	bitcoinTx := dal.BitcoinTx{
		TxID: 1, Status: dal.Pending, BtcTxID: daltest.GetHashBytes(t, "1"), Timestamp: time.Now().Unix(), Note: "",
	}

	err := db.InsertBtcTx(bitcoinTx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetBitcoinTxByTxIDAndBtcTxID(1, daltest.GetHashBytes(t, "1"))
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &bitcoinTx)
}

func Test_GetIkaSignRequestByID(t *testing.T) {
	db := daltest.InitTestDB(t)
	requests := daltest.PopulateSignRequests(t, db)

	request, err := db.GetIkaSignRequestByID(requests[0].ID)
	assert.NilError(t, err)
	assert.DeepEqual(t, *request, requests[0])
}

func Test_GetIkaTxByTxIDAndIkaTxID(t *testing.T) {
	db := daltest.InitTestDB(t)
	ikaTxs := daltest.PopulateIkaTxs(t, db)

	ikaTx, err := db.GetIkaTxByTxIDAndIkaTxID(ikaTxs[0].TxID, ikaTxs[0].IkaTxID)
	assert.NilError(t, err)
	assert.DeepEqual(t, *ikaTx, ikaTxs[0])
}

func Test_GetBitcoinTxByTxIDAndBtcTxID(t *testing.T) {
	db := daltest.InitTestDB(t)
	btcTxs := daltest.PopulateBitcoinTxs(t, db)

	btcTx, err := db.GetBitcoinTxByTxIDAndBtcTxID(btcTxs[0].TxID, btcTxs[0].BtcTxID)
	assert.NilError(t, err)
	assert.DeepEqual(t, *btcTx, btcTxs[0])
}

func Test_GetPendingIkaSignRequests(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)

	requests, err := db.GetPendingIkaSignRequests()
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 2)
}

func Test_GetSignedIkaSignRequests(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)

	signedTxs, err := db.GetSignedIkaSignRequests()
	assert.NilError(t, err)
	assert.Equal(t, len(signedTxs), 1)
}

func Test_GetBroadcastedBitcoinTxsInfo(t *testing.T) {
	db := daltest.InitTestDB(t)
	daltest.PopulateSignRequests(t, db)
	daltest.PopulateBitcoinTxs(t, db)

	signedTxs, err := db.GetBroadcastedBitcoinTxsInfo()
	assert.NilError(t, err)
	assert.Equal(t, len(signedTxs), 1)
}

func Test_UpdateIkaSignRequestFinalSig(t *testing.T) {
	db := daltest.InitTestDB(t)
	requestID := uint64(1)
	request := dal.IkaSignRequest{
		ID:        1,
		Payload:   []byte("payload"),
		DWalletID: "dwallet_id",
		UserSig:   "user_sig",
		FinalSig:  nil,
		Timestamp: time.Now().Unix(),
	}

	finalSig := []byte("final_sig1")
	err := db.InsertIkaSignRequest(request)
	assert.NilError(t, err)

	err = db.UpdateIkaSignRequestFinalSig(requestID, finalSig)
	assert.NilError(t, err)

	updatedRequest, err := db.GetIkaSignRequestByID(requestID)
	assert.NilError(t, err)
	assert.DeepEqual(t, updatedRequest.FinalSig, finalSig)
}

func Test_UpdateBitcoinTxToConfirmed(t *testing.T) {
	db := daltest.InitTestDB(t)

	bitcoinTx := dal.BitcoinTx{
		TxID: 1, Status: dal.Pending, BtcTxID: daltest.GetHashBytes(t, "1"), Timestamp: time.Now().Unix(), Note: "",
	}

	err := db.InsertBtcTx(bitcoinTx)
	assert.NilError(t, err)

	tx, err := db.GetBitcoinTxByTxIDAndBtcTxID(bitcoinTx.TxID, bitcoinTx.BtcTxID)
	assert.NilError(t, err)
	assert.Equal(t, tx.Status, dal.Pending)

	err = db.UpdateBitcoinTxToConfirmed(bitcoinTx.TxID, bitcoinTx.BtcTxID)
	assert.NilError(t, err)

	confirmedTx, err := db.GetBitcoinTxByTxIDAndBtcTxID(bitcoinTx.TxID, bitcoinTx.BtcTxID)
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, dal.Confirmed)
}
