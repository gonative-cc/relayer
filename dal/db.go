package dal

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// TxStatus represents the different states of a transaction.
type TxStatus byte

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

// NewDB creates a new DB instance
func NewDB(dbPath string) (*DB, error) {
	db := &DB{}

	var err error
	db.conn, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, err
}

// InitDB initializes the database
func (db DB) InitDB() error {
	_, err := db.conn.Exec(createTransactionsTableSQL)
	return err
}

// InsertTx inserts a new transaction into the database
func (db DB) InsertTx(tx Tx) error {
	_, err := db.conn.Exec(insertTransactionSQL, tx.BtcTxID, tx.RawTx, tx.Status)
	return err
}

// GetTx retrives a transaction by its txid
func (db DB) GetTx(txID uint64) (*Tx, error) {
	row := db.conn.QueryRow(getTransactionByTxidSQL, txID)
	var tx Tx
	err := row.Scan(&tx.BtcTxID, &tx.RawTx, &tx.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
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
	_, err := db.conn.Exec(updateTransactionStatusSQL, status, txID)
	return err
}

func (db DB) Close() error {
	return db.conn.Close()
}
