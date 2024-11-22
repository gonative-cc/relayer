package database

import (
	"testing"
)

func TestInsertTransaction(t *testing.T) {
	err := InitDB(":memory:") // in-memory database
	if err != nil {
		t.Fatal(err)
	}

	tx := Transaction{
		Txid:   "test-txid",
		RawTx:  "raw-transaction-hex",
		Status: StatusPending,
	}

	err = InsertTransaction(tx)
	if err != nil {
		t.Errorf("InsertTransaction() error = %v", err)
	}
}

func TestGetPendingTransactions(t *testing.T) {
	err := InitDB(":memory:") // in-memory database
	if err != nil {
		t.Fatal(err)
	}

	transactions := []Transaction{
		{Txid: "tx1", RawTx: "tx1-hex", Status: StatusPending},
		{Txid: "tx2", RawTx: "tx2-hex", Status: StatusBroadcasted},
		{Txid: "tx3", RawTx: "tx3-hex", Status: StatusPending},
	}
	for _, tx := range transactions {
		err = InsertTransaction(tx)
		if err != nil {
			t.Fatal(err)
		}
	}

	pendingTxs, err := GetPendingTransactions()
	if err != nil {
		t.Errorf("GetPendingTransactions() error = %v", err)
	}

	if len(pendingTxs) != 2 {
		t.Errorf("Expected 2 pending transactions, got %d", len(pendingTxs))
	}
}

func TestUpdateTransactionStatus(t *testing.T) {
	err := InitDB(":memory:") // in-memory database
	if err != nil {
		t.Fatal(err)
	}

	tx := Transaction{
		Txid:   "test-txid",
		RawTx:  "raw-transaction-hex",
		Status: StatusPending,
	}
	err = InsertTransaction(tx)
	if err != nil {
		t.Fatal(err)
	}

	err = UpdateTransactionStatus("test-txid", StatusBroadcasted)
	if err != nil {
		t.Errorf("UpdateTransactionStatus() error = %v", err)
	}

	updatedTx, err := GetTransaction("test-txid")
	if err != nil {
		t.Errorf("GetTransaction() error = %v", err)
	}

	if updatedTx.Status != StatusBroadcasted {
		t.Errorf("Expected StatusBrodcasted, got %s", updatedTx.Status)
	}
}
