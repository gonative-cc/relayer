package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// TxStatus represents the different states of a transaction.
type TxStatus int

// Transaction status constants
const (
	StatusPending TxStatus = iota
	StatusBroadcasted
	StatusConfirmed
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
            status INTEGER NOT NULL NOT NULL
        )
    `
)

// Tx represents a transaction record in the database.
type Tx struct {
	BtcTxID uint64   `json:"txid"`
	RawTx   string   `json:"rawtx"`
	Status  TxStatus `json:"status"`
	// TODO: other fields
}

// DB holds the database connection and provides methods for interacting with it.
type DB struct {
	conn *sql.DB
}

// NewDB creates a new DB instance and initializes the database connection.
func NewDB(dbPath string) (*DB, error) {
	db := &DB{} // Create a new DB instance

	// Initialize the database connection
	var err error
	db.conn, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.conn.Exec(createTransactionsTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil // Return the initialized DB instance
}

// InsertTx inserts a new transaction into the database
func (db DB) InsertTx(tx Tx) error {
	stmt, err := db.conn.Prepare(insertTransactionSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(tx.BtcTxID, tx.RawTx, tx.Status)
	if err != nil {
		return err
	}

	return nil
}

// GetTx retrives a transaction by its txid
func (db DB) GetTx(txID uint64) (*Tx, error) {
	stmt, err := db.conn.Prepare(getTransactionByTxidSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(txID)

	var tx Tx
	err = row.Scan(&tx.BtcTxID, &tx.RawTx, &tx.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if no transaction was found TODO:not sure if its ideal
		}
		return nil, err
	}

	return &tx, nil
}

// GetPendingTxs retrieves all transactions with a "pending" status
func (db DB) GetPendingTxs() ([]Tx, error) {
	rows, err := db.conn.Query(getPendingTransactionsSQL, StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Tx
	for rows.Next() {
		var tx Tx
		err := rows.Scan(&tx.BtcTxID, &tx.RawTx, &tx.Status)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// UpdateTxStatus updates the status of a transaction by txid
func (db DB) UpdateTxStatus(txID uint64, status TxStatus) error {
	stmt, err := db.conn.Prepare(updateTransactionStatusSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, txID)
	if err != nil {
		return err
	}

	return nil
}
