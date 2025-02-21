package dal_test

import (
	"context"
	"testing"
	"time"

	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/daltest"
	"gotest.tools/v3/assert"
)

func Test_InsertIkaSignRequest(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)

	request := dal.IkaSignRequest{
		ID:        1,
		Payload:   []byte("payload"),
		DWalletID: "dwallet_id",
		UserSig:   "user_sig",
		FinalSig:  nil,
		Timestamp: time.Now().Unix(),
	}

	err := db.InsertIkaSignRequest(ctx, request)
	assert.NilError(t, err)

	retrievedReq, err := db.GetIkaSignRequestByID(ctx, 1)
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedReq, &request)
}

func Test_InsertIkaTx(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)

	ikaTx := dal.IkaTx{
		SrID: 1, Status: int64(dal.Success), IkaTxID: "ika_tx_1", Timestamp: time.Now().Unix(),
	}

	err := db.InsertIkaTx(ctx, ikaTx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetIkaTx(ctx, 1, "ika_tx_1")
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &ikaTx)
}

func Test_InsertBitcoinTx(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)

	bitcoinTx := dal.BitcoinTx{
		SrID: 1, Status: int64(dal.Pending), BtcTxID: daltest.DecodeBTCHash(t, "1"), Timestamp: time.Now().Unix(),
	}

	err := db.InsertBtcTx(ctx, bitcoinTx)
	assert.NilError(t, err)

	retrievedTx, err := db.GetBitcoinTx(ctx, 1, daltest.DecodeBTCHash(t, "1"))
	assert.NilError(t, err)
	assert.DeepEqual(t, retrievedTx, &bitcoinTx)
}

func Test_GetIkaSignRequestByID(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)

	requests := daltest.PopulateSignRequests(ctx, t, db)

	request, err := db.GetIkaSignRequestByID(ctx, requests[0].ID)
	assert.NilError(t, err)
	assert.DeepEqual(t, *request, requests[0])
}

func Test_GetIkaTx(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	ikaTxs := daltest.PopulateIkaTxs(ctx, t, db)

	ikaTx, err := db.GetIkaTx(ctx, ikaTxs[0].SrID, ikaTxs[0].IkaTxID)
	assert.NilError(t, err)
	assert.DeepEqual(t, *ikaTx, ikaTxs[0])
}

func Test_GetBitcoinTx(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	btcTxs := daltest.PopulateBitcoinTxs(ctx, t, db)

	btcTx, err := db.GetBitcoinTx(ctx, btcTxs[0].SrID, btcTxs[0].BtcTxID)
	assert.NilError(t, err)
	assert.DeepEqual(t, *btcTx, btcTxs[0])
}

func Test_GetPendingIkaSignRequests(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	daltest.PopulateSignRequests(ctx, t, db)

	requests, err := db.GetPendingIkaSignRequests(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 2)
}

func Test_GetBitcoinTxsToBroadcast(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	daltest.PopulateSignRequests(ctx, t, db)
	daltest.PopulateBitcoinTxs(ctx, t, db)

	signedTxs, err := db.GetBitcoinTxsToBroadcast(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(signedTxs), 1)
}

func Test_GetBroadcastedBitcoinTxsInfo(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	daltest.PopulateSignRequests(ctx, t, db)
	daltest.PopulateBitcoinTxs(ctx, t, db)

	signedTxs, err := db.GetBroadcastedBitcoinTxsInfo(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(signedTxs), 1)
}

func Test_UpdateIkaSignRequestFinalSig(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	requestID := int64(1)
	request := dal.IkaSignRequest{
		ID:        1,
		Payload:   []byte("payload"),
		DWalletID: "dwallet_id",
		UserSig:   "user_sig",
		FinalSig:  nil,
		Timestamp: time.Now().Unix(),
	}

	finalSig := []byte("final_sig1")
	err := db.InsertIkaSignRequest(ctx, request)
	assert.NilError(t, err)

	err = db.UpdateIkaSignRequestFinalSig(ctx, requestID, finalSig)
	assert.NilError(t, err)

	updatedRequest, err := db.GetIkaSignRequestByID(ctx, requestID)
	assert.NilError(t, err)
	assert.DeepEqual(t, updatedRequest.FinalSig, finalSig)
}

func Test_UpdateBitcoinTxToConfirmed(t *testing.T) {
	ctx := context.Background()
	db := daltest.InitTestDB(t, ctx)
	bitcoinTx := dal.BitcoinTx{
		SrID: 1, Status: int64(dal.Pending), BtcTxID: daltest.DecodeBTCHash(t, "1"), Timestamp: time.Now().Unix(),
	}

	err := db.InsertBtcTx(ctx, bitcoinTx)
	assert.NilError(t, err)

	tx, err := db.GetBitcoinTx(ctx, bitcoinTx.SrID, bitcoinTx.BtcTxID)
	assert.NilError(t, err)
	assert.Equal(t, tx.Status, int64(dal.Pending))

	err = db.UpdateBitcoinTxToConfirmed(ctx, bitcoinTx.SrID, bitcoinTx.BtcTxID)
	assert.NilError(t, err)

	confirmedTx, err := db.GetBitcoinTx(ctx, bitcoinTx.SrID, bitcoinTx.BtcTxID)
	assert.NilError(t, err)
	assert.Equal(t, confirmedTx.Status, int64(dal.Confirmed))
}
