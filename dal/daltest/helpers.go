package daltest

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/assert"
)

// InitTestDB initializes an in-memory database for testing purposes.
func InitTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}

// GetHashBytes creates a byte array from a hash string.
func GetHashBytes(t *testing.T, hashString string) []byte {
	t.Helper()
	hash, err := chainhash.NewHashFromStr(hashString)
	assert.NilError(t, err)
	return hash.CloneBytes()

}

// PopulateSignRequests inserts a set of predefined IkaSignRequest into the database.
func PopulateSignRequests(t *testing.T, db *dal.DB) []dal.IkaSignRequest {
	t.Helper()

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

	requests := []dal.IkaSignRequest{
		{ID: 1, Payload: rawTxBytes, DWalletID: "dwallet1", UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()},
		{ID: 2, Payload: rawTxBytes, DWalletID: "dwallet2", UserSig: "user_sig2", FinalSig: []byte("final_sig2"), Timestamp: time.Now().Unix()},
		{ID: 3, Payload: rawTxBytes, DWalletID: "dwallet3", UserSig: "user_sig3", FinalSig: nil, Timestamp: time.Now().Unix()},
		{ID: 4, Payload: rawTxBytes, DWalletID: "dwallet4", UserSig: "user_sig4", FinalSig: []byte("final_sig4"), Timestamp: time.Now().Unix()},
	}

	for _, request := range requests {
		err := db.InsertIkaSignRequest(request)
		assert.NilError(t, err)
	}

	return requests
}

// PopulateIkaTxs inserts a set of predefined IkaTxs into the database.
func PopulateIkaTxs(t *testing.T, db *dal.DB) []dal.IkaTx {
	t.Helper()

	ikaTxs := []dal.IkaTx{
		{TxID: 1, Status: dal.Success, IkaTxID: "ika_tx_1", Timestamp: time.Now().Unix(), Note: ""},
		{TxID: 2, Status: dal.Success, IkaTxID: "ika_tx_2", Timestamp: time.Now().Unix(), Note: ""},
		{TxID: 3, Status: dal.Failed, IkaTxID: "ika_tx_3", Timestamp: time.Now().Unix(), Note: "some error"},
	}

	for _, tx := range ikaTxs {
		err := db.InsertIkaTx(tx)
		assert.NilError(t, err)
	}

	return ikaTxs
}

// PopulateBitcoinTxs inserts a set of predefined BitcoinTxs into the database.
func PopulateBitcoinTxs(t *testing.T, db *dal.DB) []dal.BitcoinTx {
	t.Helper()

	bitcoinTxs := []dal.BitcoinTx{
		{TxID: 2, Status: dal.Pending, BtcTxID: GetHashBytes(t, "1"), Timestamp: time.Now().Unix(), Note: ""},
		{TxID: 4, Status: dal.Pending, BtcTxID: GetHashBytes(t, "2"), Timestamp: time.Now().Unix(), Note: ""},
		{TxID: 4, Status: dal.Broadcasted, BtcTxID: GetHashBytes(t, "3"), Timestamp: time.Now().Unix(), Note: ""},
	}

	for _, tx := range bitcoinTxs {
		err := db.InsertBtcTx(tx)
		assert.NilError(t, err)
	}

	return bitcoinTxs
}
