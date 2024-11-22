package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// TransactionStatus represents the different states of a transaction.
type TransactionStatus string

// Different tx states
const (
	StatusPending     TransactionStatus = "pending"
	StatusBroadcasted TransactionStatus = "broadcasted"
	StatusConfirmed   TransactionStatus = "confirmed"
)

// SQL queries
const (
	insertTransactionSQL       = "INSERT INTO transactions(txid, rawtx, status) values(?,?,?)"
	getPendingTransactionsSQL  = "SELECT txid, rawtx, status FROM transactions WHERE status = ?"
	getTransactionByTxidSQL    = "SELECT txid, rawtx, status FROM transactions WHERE txid = ?"
	updateTransactionStatusSQL = "UPDATE transactions SET status = ? WHERE txid = ?"
	createTransactionsTableSQL = `
        CREATE TABLE IF NOT EXISTS transactions (
            txid TEXT PRIMARY KEY,
            rawtx TEXT NOT NULL,
            status TEXT NOT NULL
        )
    `
)

// Transaction represents a transaction record in the database.
type Transaction struct {
	BtcTxID uint64            `json:"txid"`
	RawTx   string            `json:"rawtx"`
	Status  TransactionStatus `json:"status"`
	// TODO: other fields
}

var db *sql.DB

// InitDB initializes the database connection and creates the table if it doesn't exist
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(createTransactionsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// InsertTransaction inserts a new transaction into the database
func InsertTransaction(tx Transaction) error {
	stmt, err := db.Prepare(insertTransactionSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(tx.BtcTxID, tx.RawTx, tx.Status)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

// GetTransaction retrives a transaction by its txid
func GetTransaction(txID uint64) (*Transaction, error) {
	stmt, err := db.Prepare(getTransactionByTxidSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(txID)

	var tx Transaction
	err = row.Scan(&tx.BtcTxID, &tx.RawTx, &tx.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if no transaction was found
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &tx, nil
}

// GetPendingTransactions retrieves all transactions with a "pending" status
func GetPendingTransactions() ([]Transaction, error) {
	rows, err := db.Query(getPendingTransactionsSQL, StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.BtcTxID, &tx.RawTx, &tx.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// UpdateTransactionStatus updates the status of a transaction by txid
func UpdateTransactionStatus(txID uint64, status TransactionStatus) error {
	stmt, err := db.Prepare(updateTransactionStatusSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, txID)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}
