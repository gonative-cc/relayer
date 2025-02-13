package daltest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/dal/internal"
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
func PopulateSignRequests(ctx context.Context, t *testing.T, db dal.DB) []internal.IkaSignRequest {
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

	requests := []internal.IkaSignRequest{
		{ID: 1, Payload: rawTxBytes, DwalletID: "dwallet1",
			UserSig: "user_sig1", FinalSig: nil, Timestamp: time.Now().Unix()},
		{ID: 2, Payload: rawTxBytes, DwalletID: "dwallet2",
			UserSig: "user_sig2", FinalSig: []byte("final_sig2"), Timestamp: time.Now().Unix()},
		{ID: 3, Payload: rawTxBytes, DwalletID: "dwallet3",
			UserSig: "user_sig3", FinalSig: nil, Timestamp: time.Now().Unix()},
		{ID: 4, Payload: rawTxBytes, DwalletID: "dwallet4",
			UserSig: "user_sig4", FinalSig: []byte("final_sig4"), Timestamp: time.Now().Unix()},
	}

	for _, request := range requests {
		err := db.InsertIkaSignRequest(ctx, request)
		assert.NilError(t, err)
	}

	return requests
}

// PopulateIkaTxs inserts a set of predefined IkaTxs into the database.
func PopulateIkaTxs(ctx context.Context, t *testing.T, db dal.DB) []internal.IkaTx {
	t.Helper()

	ikaTxs := []internal.IkaTx{
		{SrID: 1, Status: int64(dal.Success), IkaTxID: "ika_tx_1", Timestamp: time.Now().Unix()},
		{SrID: 2, Status: int64(dal.Success), IkaTxID: "ika_tx_2", Timestamp: time.Now().Unix()},
		{SrID: 3, Status: int64(dal.Failed), IkaTxID: "ika_tx_3", Timestamp: time.Now().Unix()},
	}

	for _, tx := range ikaTxs {
		err := db.InsertIkaTx(ctx, tx)
		assert.NilError(t, err)
	}

	return ikaTxs
}

// PopulateBitcoinTxs inserts a set of predefined BitcoinTxs into the database.
func PopulateBitcoinTxs(ctx context.Context, t *testing.T, db dal.DB) []internal.BitcoinTx {
	t.Helper()

	bitcoinTxs := []internal.BitcoinTx{
		{SrID: 2, Status: int64(dal.Pending), BtcTxID: DecodeBTCHash(t, "1"), Timestamp: time.Now().Unix()},
		{SrID: 4, Status: int64(dal.Pending), BtcTxID: DecodeBTCHash(t, "2"), Timestamp: time.Now().Unix()},
		{SrID: 4, Status: int64(dal.Broadcasted), BtcTxID: DecodeBTCHash(t, "3"), Timestamp: time.Now().Unix()},
	}

	for _, tx := range bitcoinTxs {
		err := db.InsertBtcTx(ctx, tx)
		assert.NilError(t, err)
	}

	return bitcoinTxs
}

// PopulateDB inserts a set of predefined data to all the tables.
func PopulateDB(ctx context.Context, t *testing.T, db dal.DB) {
	t.Helper()
	PopulateBitcoinTxs(ctx, t, db)
	PopulateIkaTxs(ctx, t, db)
	PopulateSignRequests(ctx, t, db)
}
