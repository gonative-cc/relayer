package daltest

import (
	"testing"

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

// PopulateDB inserts a set of predefined transactions into the database.
func PopulateDB(t *testing.T, db *dal.DB) []dal.Tx {
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

	txs := []dal.Tx{
		{BtcTxID: 1, RawTx: rawTxBytes, Status: dal.StatusBroadcasted, Hash: GetHashBytes(t, "1")},
		{BtcTxID: 2, RawTx: rawTxBytes, Status: dal.StatusBroadcasted, Hash: GetHashBytes(t, "2")},
		{BtcTxID: 3, RawTx: rawTxBytes, Status: dal.StatusSigned, Hash: GetHashBytes(t, "3")},
		{BtcTxID: 4, RawTx: rawTxBytes, Status: dal.StatusSigned, Hash: GetHashBytes(t, "4")},
	}
	for _, tx := range txs {
		err := db.InsertTx(tx)
		assert.NilError(t, err)
	}

	return txs
}

// PopulateNativeDB inserts a set of predefined native transactions into the database.
func PopulateNativeDB(t *testing.T, db *dal.DB) []dal.NativeTx {
	t.Helper()
	messages := [][]byte{[]byte("message1"), []byte("message2")}
	nativeTxs := []dal.NativeTx{
		{TxID: 1, DWalletCapID: "dwallet1", SignMessagesID: "sign1", Messages: messages, Status: dal.NativeTxStatusPending},
		{TxID: 2, DWalletCapID: "dwallet2", SignMessagesID: "sign2", Messages: messages, Status: dal.NativeTxStatusPending},
		{TxID: 3, DWalletCapID: "dwallet3", SignMessagesID: "sign3", Messages: messages, Status: dal.NativeTxStatusProcessed},
	}

	for _, tx := range nativeTxs {
		err := db.InsertNativeTx(tx)
		assert.NilError(t, err)
	}
	return nativeTxs
}
