package daltest

import (
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/assert"
)

// Some tests are run in parallel. The sql.Open("sqlite3", ":memory:") opens a connection for the
// in-memory DB. The DB is destroyed only after the last connection is properly closed. Trying
// to open a new connection when the previous in-memory DB is not yet closed causes error:
// according to the spec [1]: "then that database always has a private cache and is only visible
// to the database connection that originally opened it." To allow multiple connections we
// can add shared cash option, example:
//
//	rc = sqlite3_open("file::memory:?cache=shared", &db);
//
// However, in tests, we want to have a clear state each time we start a new test. So we want to
// assure a new DB is created. We do it by specify a counter for DB.
//
// [1] https://www.sqlite.org/inmemorydb.html
var testDBCounter = 0

// InitTestDB initializes an in-memory database for testing purposes. Subsequent calls will
// create a new in-memory DB.
func InitTestDB(t *testing.T) dal.DB {
	t.Helper()

	testDBCounter++
	db, err := dal.NewDB(fmt.Sprintf("file:db%d?mode=memory&cache=shared", testDBCounter))
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}

// DecodeBTCHash decodes a byte-reversed hexadecimal string.
func DecodeBTCHash(t *testing.T, hashString string) []byte {
	t.Helper()
	hash, err := chainhash.NewHashFromStr(hashString)
	assert.NilError(t, err)
	return hash.CloneBytes()

}

// PopulateSignRequests inserts a set of predefined IkaSignRequest into the database.
func PopulateSignRequests(t *testing.T, db dal.DB) []dal.IkaSignRequest {
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
		{ID: 1, Payload: rawTxBytes, DWalletID: "dwallet1",
			UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()},
		{ID: 2, Payload: rawTxBytes, DWalletID: "dwallet2",
			UserSig: "user_sig2", FinalSig: []byte("final_sig2"), Timestamp: time.Now().Unix()},
		{ID: 3, Payload: rawTxBytes, DWalletID: "dwallet3",
			UserSig: "user_sig3", FinalSig: nil, Timestamp: time.Now().Unix()},
		{ID: 4, Payload: rawTxBytes, DWalletID: "dwallet4",
			UserSig: "user_sig4", FinalSig: []byte("final_sig4"), Timestamp: time.Now().Unix()},
	}

	for _, request := range requests {
		err := db.InsertIkaSignRequest(request)
		assert.NilError(t, err)
	}

	return requests
}

// PopulateIkaTxs inserts a set of predefined IkaTxs into the database.
func PopulateIkaTxs(t *testing.T, db dal.DB) []dal.IkaTx {
	t.Helper()

	ikaTxs := []dal.IkaTx{
		{SrID: 1, Status: dal.Success, IkaTxID: "ika_tx_1", Timestamp: time.Now().Unix(), Note: ""},
		{SrID: 2, Status: dal.Success, IkaTxID: "ika_tx_2", Timestamp: time.Now().Unix(), Note: ""},
		{SrID: 3, Status: dal.Failed, IkaTxID: "ika_tx_3", Timestamp: time.Now().Unix(), Note: "some error"},
	}

	for _, tx := range ikaTxs {
		err := db.InsertIkaTx(tx)
		assert.NilError(t, err)
	}

	return ikaTxs
}

// PopulateBitcoinTxs inserts a set of predefined BitcoinTxs into the database.
func PopulateBitcoinTxs(t *testing.T, db dal.DB) []dal.BitcoinTx {
	t.Helper()

	bitcoinTxs := []dal.BitcoinTx{
		{SrID: 2, Status: dal.Pending, BtcTxID: DecodeBTCHash(t, "1"), Timestamp: time.Now().Unix(), Note: ""},
		{SrID: 4, Status: dal.Pending, BtcTxID: DecodeBTCHash(t, "2"), Timestamp: time.Now().Unix(), Note: ""},
		{SrID: 4, Status: dal.Broadcasted, BtcTxID: DecodeBTCHash(t, "3"), Timestamp: time.Now().Unix(), Note: ""},
	}

	for _, tx := range bitcoinTxs {
		err := db.InsertBtcTx(tx)
		assert.NilError(t, err)
	}

	return bitcoinTxs
}

// PopulateDB inserts a set of predefined data to all the tables.
func PopulateDB(t *testing.T, db dal.DB) {
	t.Helper()
	PopulateBitcoinTxs(t, db)
	PopulateIkaTxs(t, db)
	PopulateSignRequests(t, db)
}
