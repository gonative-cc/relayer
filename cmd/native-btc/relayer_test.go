package main

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/native/database"
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
	err := database.InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	config := Config{
		DatabasePath: ":memory:",
		BitcoinNode:  "mock-node",
	}
	relayer, err := NewRelayer(config)
	if err != nil {
		t.Fatal(err)
	}
	relayer.btcClient = &MockBitcoinClient{}

	transactions := []database.Transaction{
		{BtcTxID: 1, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0100f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: database.StatusPending},
		{BtcTxID: 2, RawTx: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0200f2052a010000001976a914000000000000000000000000000000000000000088ac00000000", Status: database.StatusPending},
	}
	for _, tx := range transactions {
		err = database.InsertTransaction(tx)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Start the relayer in a separate goroutine
	go relayer.Start()

	time.Sleep(time.Second * 2)

	for _, tx := range transactions {
		updatedTx, err := database.GetTransaction(tx.BtcTxID)
		if err != nil {
			t.Errorf("Error getting transaction: %v", err)
		}
		if updatedTx.Status != database.StatusBroadcasted {
			t.Errorf("Expected transaction status to be '%s', but got '%s'", database.StatusBroadcasted, updatedTx.Status)
		}
	}

	relayer.Stop()
}
