package dal

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// IkaSignRequest represents a row in the `ika_sign_requests` table.
type IkaSignRequest struct {
	ID        uint64    `json:"id"`
	Payload   Payload   `json:"payload"`
	DWalletID string    `json:"dwallet_id"`
	UserSig   string    `json:"user_sig"`
	FinalSig  Signature `json:"final_sig"`
	Timestamp uint64    `json:"time"`
}

// IkaTx represents a row in the `ika_txs` table.
type IkaTx struct {
	TxID      uint64      `json:"tx_id"`
	Status    IkaTxStatus `json:"status"`
	IkaTxID   string      `json:"ika_tx_id"`
	Timestamp uint64      `json:"time"`
	Note      string      `json:"note"`
}

// BitcoinTx represents a row in the `bitcoin_txs` table.
type BitcoinTx struct {
	TxID      uint64          `json:"tx_id"`
	Status    BitcoinTxStatus `json:"status"`
	BtcTxId   []byte          `json:"btc_tx_id"`
	Timestamp uint64          `json:"time"`
	Note      string          `json:"note"`
}

// IkaTxStatus represents the different states of a native transaction.
type IkaTxStatus byte

// Native ransaction status constants
const (
	Success IkaTxStatus = iota
	Failed
)

// BitcoinTxStatus represents the different states of a bitcoin transaction.
type BitcoinTxStatus byte

// Native ransaction status constants
const (
	Pending BitcoinTxStatus = iota
	Broadcasted
	Confirmed
)

// Payload is an alias for []byte, representing a single payload to be singed.
type Payload = []byte

// Signature is an alias for []byte, representing the final signature.
type Signature = []byte

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

// InitDB initializes the database and creates the tables.
func (db *DB) InitDB() error {
	const createIkaSignRequestsTableSQL = `
		CREATE TABLE IF NOT EXISTS ika_sign_requests (
			id INTEGER PRIMARY KEY,
			payload BLOB NOT NULL,
			dwallet_id TEXT NOT NULL,
			user_sig TEXT NOT NULL,
			final_sig BLOB,
			timestamp INTEGER
		)`
	const createIkaTxsTableSQL = `
		CREATE TABLE IF NOT EXISTS ika_txs (
			tx_id INTEGER NOT NULL,
			status INTEGER NOT NULL,
			ika_tx_id TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			note TEXT,
			PRIMARY KEY (tx_id, ika_tx_id),
			FOREIGN KEY (tx_id) REFERENCES ika_sign_requests (id)
		)`

	const createBitcoinTxsTableSQL = `
		CREATE TABLE IF NOT EXISTS bitcoin_txs (
			tx_id INTEGER NOT NULL,
			status INTEGER NOT NULL,
			btc_tx_id BlOB NOT NULL,
			timestamp INTEGER NOT NULL,
			note TEXT,
			PRIMARY KEY (tx_id, btc_tx_id),
			FOREIGN KEY (tx_id) REFERENCES ika_sign_requests (id)
		)`
	_, err := db.conn.Exec(createIkaSignRequestsTableSQL)
	if err != nil {
		return fmt.Errorf("creating ika_sign_requests table: %w", err)
	}
	_, err = db.conn.Exec(createIkaTxsTableSQL)
	if err != nil {
		return fmt.Errorf("creating ika_txs table: %w", err)
	}
	_, err = db.conn.Exec(createBitcoinTxsTableSQL)
	if err != nil {
		return fmt.Errorf("creating bitcoin_txs table: %w", err)
	}

	return nil
}

// InsertIkaSignRequest inserts a new transaction into the database
func (db *DB) InsertIkaSignRequest(signReq IkaSignRequest) error {
	const insertIkaSignRequestSQL = `
		INSERT INTO ika_sign_requests (id, payload, dwallet_id, user_sig, timestamp) 
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertIkaSignRequestSQL, signReq.ID, signReq.Payload, signReq.DWalletID, signReq.UserSig, signReq.Timestamp)
	return err
}

// InsertIkaTx inserts a new Ika transaction into the database.
func (db *DB) InsertIkaTx(tx IkaTx) error {
	const insertIkaTxSQL = `
		INSERT INTO ika_txs (tx_id, status, ika_tx_id, timestamp, note) 
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertIkaTxSQL, tx.TxID, tx.Status, tx.IkaTxID, tx.Timestamp, tx.Note)
	return err
}

// InsertBtcTx inserts a new Bitcoin transaction into the database.
func (db *DB) InsertBtcTx(tx BitcoinTx) error {
	const insertBitcoinTxSQL = `
		INSERT INTO bitcoin_txs (tx_id, status, btc_tx_id, timestamp, note) 
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertBitcoinTxSQL, tx.TxID, tx.Status, tx.BtcTxId, tx.Timestamp, tx.Note)
	return err
}

// GetIkaSignRequest retrives a signature request by its id
func (db DB) GetIkaSignRequest(id uint64) (*IkaSignRequest, error) {
	const getIkaSignRequestByIdSQL = `
		SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
		FROM ika_sign_requests
		WHERE id = ?`
	row := db.conn.QueryRow(getIkaSignRequestByIdSQL, id)
	var signReq IkaSignRequest
	err := row.Scan(&signReq.ID, &signReq.Payload, &signReq.DWalletID, &signReq.UserSig, &signReq.FinalSig, &signReq.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &signReq, nil
}

// GetIkaTxByTxIdAndIkaTxId retrieves an Ika transaction by its primary key (tx_id and ika_tx_id).
func (db *DB) GetIkaTxByTxIdAndIkaTxId(txID uint64, ikaTxID string) (*IkaTx, error) {
	row := db.conn.QueryRow(`
        SELECT tx_id, status, ika_tx_id, timestamp, note
        FROM ika_txs
        WHERE tx_id = ? AND ika_tx_id = ?`,
		txID, ikaTxID)

	var ikaTx IkaTx
	err := row.Scan(&ikaTx.TxID, &ikaTx.Status, &ikaTx.IkaTxID, &ikaTx.Timestamp, &ikaTx.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // no rows
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &ikaTx, nil
}

// GetBitcoinTxByTxIDAndBtcTxID retrieves a Bitcoin transaction by its primary key (tx_id and btc_tx_id).
func (db *DB) GetBitcoinTxByTxIDAndBtcTxID(txID uint64, btcTxId string) (*BitcoinTx, error) {
	row := db.conn.QueryRow(`
        SELECT tx_id, status, btc_tx_id, timestamp, note
        FROM bitcoin_txs
        WHERE tx_id = ? AND btc_tx_id = ?`,
		txID, btcTxId)

	var bitcoinTx BitcoinTx
	err := row.Scan(&bitcoinTx.TxID, &bitcoinTx.Status, &bitcoinTx.BtcTxId, &bitcoinTx.Timestamp, &bitcoinTx.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // no rows
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &bitcoinTx, nil
}

// GetIkaSignRequestWithStatus retrieves an IkaSignRequest with its associated IkaTx status.
func (db *DB) GetIkaSignRequestWithStatus(id uint64) (*IkaSignRequest, IkaTxStatus, error) {
	// Use a JOIN query to efficiently retrieve the data from both tables
	row := db.conn.QueryRow(`
        SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp, it.status
        FROM ika_sign_requests sr
        INNER JOIN ika_txs it ON sr.id = it.tx_id
        WHERE sr.id = ?
        ORDER BY it.time DESC -- Get the latest status
        LIMIT 1
    `, id)

	var request IkaSignRequest
	var status IkaTxStatus
	err := row.Scan(&request.ID, &request.Payload, &request.DWalletID, &request.UserSig, &request.FinalSig, &request.Timestamp, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to scan row: %w", err)
	}

	return &request, status, nil
}

// GetSignedIkaSignRequests retrieves IkaSignRequests with BitcoinTxStatus "Pending".
func (db *DB) GetSignedIkaSignRequests() ([]IkaSignRequest, error) {
	return db.getIkaSignRequestsWithBtcTxStatus(Pending)
}

// GetBroadcastedIkaSignRequests retrieves IkaSignRequests with BitcoinTxStatus "Broadcasted".
func (db *DB) GetBroadcastedIkaSignRequests() ([]IkaSignRequest, error) {
	return db.getIkaSignRequestsWithBtcTxStatus(Broadcasted)
}

// getIkaSignRequestsWithBtcTxStatus retrieves IkaSignRequests with the given BitcoinTxStatus.
func (db *DB) getIkaSignRequestsWithBtcTxStatus(status BitcoinTxStatus) ([]IkaSignRequest, error) {
	rows, err := db.conn.Query(`
        SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp
        FROM ika_sign_requests sr
        INNER JOIN bitcoin_txs bt ON sr.id = bt.tx_id
        WHERE bt.status = ?
    `, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var requests []IkaSignRequest
	for rows.Next() {
		var request IkaSignRequest
		err := rows.Scan(&request.ID, &request.Payload, &request.DWalletID, &request.UserSig, &request.FinalSig, &request.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// / GetPendingIkaSignRequests retrieves IkaSignRequests that need to be signed.
func (db *DB) GetPendingIkaSignRequests() ([]IkaSignRequest, error) {
	rows, err := db.conn.Query(`
        SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
        FROM ika_sign_requests
        WHERE final_sig IS NULL
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var requests []IkaSignRequest
	for rows.Next() {
		var request IkaSignRequest
		err := rows.Scan(&request.ID, &request.Payload, &request.DWalletID, &request.UserSig, &request.FinalSig, &request.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// Close closes the db connection
func (db DB) Close() error {
	return db.conn.Close()
}
