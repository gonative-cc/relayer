package dal

import (
	"context"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
)

const (
	batch         int64 = 1000
	workers       int64 = 5
	testTimestamp int64 = 123
)

func Test_DbRace(t *testing.T) {
	db, err := NewDB(":memory:")
	assert.NilError(t, err)
	ctx := context.Background()
	err = db.InitDB(ctx)
	assert.NilError(t, err)

	// Test 1: Parallel inserts
	var wg sync.WaitGroup
	for i := int64(0); i < workers; i++ {
		wg.Add(1)
		go insertManySignReq(ctx, t, &wg, db, i*batch, (i+1)*batch)
	}
	wg.Wait()

	var srID int64 = 1
	sr, err := db.GetIkaSignRequestByID(ctx, srID)
	assert.NilError(t, err)
	assert.Assert(t, sr != nil)
	assert.Equal(t, testTimestamp, sr.Timestamp)

	row := db.conn.QueryRow(`SELECT COUNT(*) FROM ika_sign_requests`)
	var count int64
	err = row.Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, batch*workers, count)

	// Test 2: Parallel updates
	for i := int64(0); i < workers; i++ {
		wg.Add(1)
		go loopIncrementIkaSRTimestamp(t, &wg, db, batch, srID)
	}
	wg.Wait()

	sr, err = db.GetIkaSignRequestByID(ctx, srID)
	assert.NilError(t, err)
	assert.Assert(t, sr != nil)
	assert.Equal(t, int64(testTimestamp+workers*batch), sr.Timestamp)

}

func insertManySignReq(ctx context.Context, t *testing.T, wg *sync.WaitGroup, db DB, idFrom, idTo int64) {
	for i := idFrom; i < idTo; i++ {
		sr := IkaSignRequest{
			ID:        int64(i),
			Payload:   []byte{},
			DWalletID: "",
			UserSig:   "",
			FinalSig:  nil,
			Timestamp: testTimestamp,
		}
		err := db.InsertIkaSignRequest(ctx, sr)
		assert.NilError(t, err)
		t.Logf("Worker %d finished inserting %d sign requests (from ID %d to %d)", i, idTo-idFrom, idFrom, idTo)
	}
	wg.Done()
}

func loopIncrementIkaSRTimestamp(t *testing.T, wg *sync.WaitGroup, db DB, n int64, srID int64) {
	for i := int64(0); i < n; i++ {
		assert.NilError(t, db.incrementIkaSRTimestamp(srID))
	}
	wg.Done()
}

func (db DB) incrementIkaSRTimestamp(srID int64) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	row := db.conn.QueryRow(`SELECT timestamp FROM ika_sign_requests WHERE id =?`, srID)
	var tm int64
	if err := row.Scan(&tm); err != nil {
		return err
	}

	_, err := db.conn.Exec(`UPDATE ika_sign_requests SET timestamp =? WHERE id =?`, tm+1, srID)
	return err
}
