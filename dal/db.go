package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// TxStatus represents the different states of a transaction.
type TxStatus byte

// Transaction status constants
const (
	StatusPending TxStatus = iota
	StatusSigned
	StatusBroadcasted
	StatusConfirmed
)

// SQL queries
const (
	insertTxSQL       = "INSERT INTO transactions(txid, hash, rawtx, status) values(?,?,?,?)"
	getTxsByStatusSQL = "SELECT txid, hash, rawtx, status FROM transactions WHERE status = ?"
	getTxByTxidSQL    = "SELECT txid, hash, rawtx, status FROM transactions WHERE txid = ?"
	updateTxStatusSQL = "UPDATE transactions SET status = ? WHERE txid = ?"
	createTxsTableSQL = `
        CREATE TABLE IF NOT EXISTS transactions (
            txid TEXT PRIMARY KEY,
			      hash BLOB NOT NULL,
            rawtx BLOB NOT NULL,
            status INTEGER NOT NULL NOT NULL
        )
    `
	createNativeTxsTableSQL = `
		CREATE TABLE IF NOT EXISTS native_transactions (
			txid INTEGER PRIMARY KEY,
			dwallet_cap_id TEXT NOT NULL,
			sign_messages_id TEXT NOT NULL,
			messages BLOB NOT NULL,
			status INTEGER NOT NULL DEFAULT 0
		)
	`
	InsertNativeTx          = "INSERT INTO native_transactions(txid, dwallet_cap_id, sign_messages_id, messages, status) VALUES(?,?,?,?,?)"
	updateNativeTxStatusSQL = "UPDATE native_transactions SET status = ? WHERE txid = ?"
	getNativeTxsByStatusSQL = "SELECT txid, dwallet_cap_id, sign_messages_id, messages, status FROM native_transactions WHERE status = ?"
	getNativeTxByTxidSQL    = "SELECT txid, dwallet_cap_id, sign_messages_id, messages, status FROM native_transactions WHERE txid = ?"
)

type NativeTxStatus int

const (
	NativeTxStatusPending   NativeTxStatus = 0
	NativeTxStatusProcessed NativeTxStatus = 1
)

// Message is an alias for []byte, representing a single message.
type Message = []byte

// NativeTx represents the data extracted from a Native chain transaction.
type NativeTx struct {
	TxID           uint64         `json:"tx_id"` // TODO: do we need multiple TxIDs here?
	DWalletCapID   string         `json:"dwallet_cap_id"`
	SignMessagesID string         `json:"sign_messages_id"`
	Messages       []Message      `json:"messages"`
	Status         NativeTxStatus `json:"status"`
}

// Tx represents a transaction record in the database.
type Tx struct {
	BtcTxID uint64   `json:"txid"`
	Hash    []byte   `json:"hash"`
	RawTx   []byte   `json:"rawtx"`
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
func (db *DB) InitDB() error {
	_, err := db.conn.Exec(createTxsTableSQL)
	if err != nil {
		return fmt.Errorf("creating transactions table: %w", err)
	}
	_, err = db.conn.Exec(createNativeTxsTableSQL)
	if err != nil {
		return fmt.Errorf("creating native_transactions table: %w", err)
	}
	return nil
}

// InsertTx inserts a new transaction into the database
func (db *DB) InsertTx(tx Tx) error {
	_, err := db.conn.Exec(insertTxSQL, tx.BtcTxID, tx.Hash, tx.RawTx, tx.Status)
	return err
}

// InsertNativeTx inserts a new native transaction.
func (db *DB) InsertNativeTx(tx NativeTx) error {
	messagesBytes, err := json.Marshal(tx.Messages)
	if err != nil {
		return fmt.Errorf("marshaling messages: %w", err)
	}
	_, err = db.conn.Exec(InsertNativeTx, tx.TxID, tx.DWalletCapID, tx.SignMessagesID, messagesBytes, tx.Status)
	return err
}

// GetTx retrives a transaction by its txid
func (db DB) GetTx(txID uint64) (*Tx, error) {
	row := db.conn.QueryRow(getTxByTxidSQL, txID)
	var tx Tx
	err := row.Scan(&tx.BtcTxID, &tx.Hash, &tx.RawTx, &tx.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tx, nil
}

// GetNativeTx retrieves a native transaction by ID.
func (db *DB) GetNativeTx(txID uint64) (*NativeTx, error) {
	row := db.conn.QueryRow(getNativeTxByTxidSQL, txID)
	var tx NativeTx
	var messagesBytes []byte
	err := row.Scan(&tx.TxID, &tx.DWalletCapID, &tx.SignMessagesID, &messagesBytes, &tx.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal(messagesBytes, &tx.Messages)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling messages: %w", err)
	}
	return &tx, nil
}

// GetSignedTxs retrieves all transactions with a "signed" status
func (db DB) GetSignedTxs() ([]Tx, error) {
	return db.getTxsByStatus(StatusSigned)
}

// GetBroadcastedTxs retrieves all transactions with a "broadcasted" status
func (db DB) GetBroadcastedTxs() ([]Tx, error) {
	return db.getTxsByStatus(StatusBroadcasted)
}

// getTxsByStatus retrieves all transactions with a given status
func (db DB) getTxsByStatus(status TxStatus) ([]Tx, error) {
	rows, err := db.conn.Query(getTxsByStatusSQL, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Tx
	for rows.Next() {
		var tx Tx
		err := rows.Scan(&tx.BtcTxID, &tx.Hash, &tx.RawTx, &tx.Status)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetNativeTxsByStatus retrieves all native transactions with a given status.
func (db *DB) GetNativeTxsByStatus(status NativeTxStatus) ([]NativeTx, error) {
	rows, err := db.conn.Query(getNativeTxsByStatusSQL, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []NativeTx
	var messagesBytes []byte
	for rows.Next() {
		var tx NativeTx
		err := rows.Scan(&tx.TxID, &tx.DWalletCapID, &tx.SignMessagesID, &messagesBytes, &tx.Status)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(messagesBytes, &tx.Messages)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling messages: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// UpdateTxStatus updates the status of a transaction by txid
func (db *DB) UpdateTxStatus(txID uint64, status TxStatus) error {
	_, err := db.conn.Exec(updateTxStatusSQL, status, txID)
	return err
}

// UpdateNativeTxStatus updates the status of a native transaction by txid
func (db *DB) UpdateNativeTxStatus(txID uint64, status NativeTxStatus) error {
	_, err := db.conn.Exec(updateNativeTxStatusSQL, status, txID)
	return err
}

// Close closes the db connection
func (db DB) Close() error {
	return db.conn.Close()
}
