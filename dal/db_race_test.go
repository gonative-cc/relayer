package dal

import (
	"testing"

	"gotest.tools/v3/assert"
)

func insertManySignReq(t *testing.T, done chan<- bool, db *DB, idFrom, idTo uint64) {
	for i := idFrom; i < idTo; i++ {
		sr := IkaSignRequest{
			ID:        i,
			Payload:   []byte{},
			DWalletID: "",
			UserSig:   "",
			FinalSig:  nil,
			Timestamp: 123,
		}
		err := db.InsertIkaSignRequest(sr)
		assert.NilError(t, err)
	}

	t.Logf("finished from %d, to: %d", idFrom, idTo)
	done <- true
}

func Test_DbRace(t *testing.T) {
	db, err := NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)

	const batch = 1000
	const workers = 5
	done := make(chan bool, workers)

	for i := uint64(0); i < workers; i++ {
		// TODO: use WorkGroup
		go insertManySignReq(t, done, db, i*batch, (i+1)*batch)
	}
	for i := 0; i < workers; i++ {
		<-done
	}

	sr, err := db.GetIkaSignRequestByID(1)
	assert.NilError(t, err)
	assert.Assert(t, sr != nil)
	assert.Equal(t, int64(123), sr.Timestamp)

	row := db.conn.QueryRow(`
SELECT COUNT(*) FROM ika_sign_requests`)
	var count int
	err = row.Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, batch*workers, count)

	// TODO: use case to do race condition for update
}
