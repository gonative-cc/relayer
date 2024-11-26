package main

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/database"
	"gotest.tools/assert"
)

// Mock Bitcoin RPC client for testing
type MockBitcoinClient struct{}

func (m *MockBitcoinClient) SendRawTransaction(_ *wire.MsgTx, _ bool) (*chainhash.Hash, error) {
	// Mock implementation
	hash, err := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (m *MockBitcoinClient) Shutdown() {
}

func TestRelayerStart(t *testing.T) {
	db := initTestDB(t)

	config := Config{
		DatabasePath: ":memory:",
		BitcoinNode:  "mock-node",
	}
	relayer, err := NewRelayer(config, db)
	assert.NilError(t, err)
	relayer.btcClient = &MockBitcoinClient{}

	transactions := []database.Tx{
		{BtcTxID: 1, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: database.StatusPending},
		{BtcTxID: 2, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0200f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: database.StatusPending},
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
		assert.Equal(t, updatedTx.Status, database.StatusBroadcasted)
	}

	relayer.Stop()
}

func initTestDB(t *testing.T) *database.DB {
	t.Helper()

	db, err := database.NewDB(":memory:")
	assert.NilError(t, err)
	return db
}
