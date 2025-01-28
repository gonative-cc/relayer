package dal

import (
	"testing"

	"gotest.tools/v3/assert"
)

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

	var srID uint64 = 1
	sr, err := db.GetIkaSignRequestByID(srID)
	assert.NilError(t, err)
	assert.Assert(t, sr != nil)
	assert.Equal(t, int64(123), sr.Timestamp)

	row := db.conn.QueryRow(`
SELECT COUNT(*) FROM ika_sign_requests`)
	var count int
	err = row.Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, batch*workers, count)

	// test 2
	// try parallel update

	for i := uint64(0); i < workers; i++ {
		go loopIncrementIkaSRTimestamp(t, done, db, batch, srID)
	}
	for i := 0; i < workers; i++ {
		<-done
	}

	sr, err = db.GetIkaSignRequestByID(srID)
	assert.NilError(t, err)
	assert.Assert(t, sr != nil)
	assert.Equal(t, int64(123+workers*batch), sr.Timestamp)

}
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

func loopIncrementIkaSRTimestamp(t *testing.T, done chan<- bool, db *DB, n int, srID uint64) {
	for i := 0; i < n; i++ {
		db.incrementIkaSRTimestamp(t, srID)
	}
	done <- true
}

func (db *DB) incrementIkaSRTimestamp(t *testing.T, srID uint64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	row := db.conn.QueryRow(`
SELECT timestamp FROM ika_sign_requests WHERE id = ?`, srID)
	var tm int64
	assert.NilError(t, row.Scan(&tm))

	_, err := db.conn.Exec(`
UPDATE ika_sign_requests
SET timestamp = ?
WHERE id = ?`, tm+1, srID)
	assert.NilError(t, err)
}
