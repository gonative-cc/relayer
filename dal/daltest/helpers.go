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
func GetHashBytes(t *testing.T, hashString tring) []byte {
	t.Helper()
	hash, err := chainhash.NewHashFromStr(hashString)
	assert.NilError(t, err)
	return hash.CloneBytes()

}

// PopulateDB inserts a set of predefined transactions into the database.
func PopulateDB(t *testing.T, db *dal.DB) []dal.Tx {
	t.Helper()

	txs := []dal.Tx{
		{BtcTxID: 1, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000001", Status: dal.StatusBroadcasted, Hash: GetHashBytes(t, "1")},
		{BtcTxID: 2, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000002", Status: dal.StatusBroadcasted, Hash: GetHashBytes(t, "2")},
		{BtcTxID: 3, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000003", Status: dal.StatusPending, Hash: GetHashBytes(t, "3")},
		{BtcTxID: 4, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000004", Status: dal.StatusPending, Hash: GetHashBytes(t, "4")},
	}
	for _, tx := range txs {
		err := db.InsertTx(tx)
		assert.NilError(t, err)
	}

	return txs
}
