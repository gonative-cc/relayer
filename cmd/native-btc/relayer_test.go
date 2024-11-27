package main

import (
	"testing"
	"time"

	"github.com/gonative-cc/relayer/bitcoin"
	"github.com/gonative-cc/relayer/dal"
	"gotest.tools/assert"
)

func TestRelayerStart(t *testing.T) {
	db := initTestDB(t)

	config := Config{
		DatabasePath: ":memory:",
		BitcoinNode:  "mock-node",
	}
	relayer, err := NewRelayer(config, db)
	assert.NilError(t, err)
	relayer.btcClient = &bitcoin.MockClient{}

	transactions := []dal.Tx{
		{BtcTxID: 1, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: dal.StatusPending},
		{BtcTxID: 2, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0200f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: dal.StatusPending},
	}
	for _, tx := range transactions {
		err = db.InsertTx(tx)
		if err != nil {
			assert.NilError(t, err)
		}
	}

	// Start the relayer in a separate goroutine
	go relayer.Start()

	time.Sleep(time.Second * 2)

	for _, tx := range transactions {
		updatedTx, err := db.GetTx(tx.BtcTxID)
		assert.NilError(t, err, "Error getting transaction")
		assert.Equal(t, updatedTx.Status, dal.StatusBroadcasted)
	}

	relayer.Stop()
}

func initTestDB(t *testing.T) *dal.DB {
	t.Helper()

	db, err := dal.NewDB(":memory:")
	assert.NilError(t, err)
	err = db.InitDB()
	assert.NilError(t, err)
	return db
}
