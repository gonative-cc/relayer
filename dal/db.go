package dal

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// IkaSignRequest represents a row in the `ika_sign_requests` table.
type IkaSignRequest struct {
	// ID is a sign request reported and managed by Native
	ID        uint64    `json:"id"`
	Payload   Payload   `json:"payload"`
	DWalletID string    `json:"dwallet_id"`
	UserSig   string    `json:"user_sig"`
	FinalSig  Signature `json:"final_sig"`
	Timestamp int64     `json:"time"`
}

// IkaTx represents a row in the `ika_txs` table.
type IkaTx struct {
	// SrID is the IkaSignRequest.ID
	SrID      uint64      `json:"sr_id"`
	Status    IkaTxStatus `json:"status"`
	IkaTxID   string      `json:"ika_tx_id"`
	Timestamp int64       `json:"timestamp"`
	Note      string      `json:"note"`
}

// BitcoinTx represents a row in the `bitcoin_txs` table.
type BitcoinTx struct {
	// SrID is the IkaSignRequest.ID
	SrID      uint64          `json:"sr_id"`
	Status    BitcoinTxStatus `json:"status"`
	BtcTxID   []byte          `json:"btc_tx_id"`
	Timestamp int64           `json:"timestamp"`
	Note      string          `json:"note"`
}

// BitcoinTxInfo holds the relevant information for a Bitcoin transaction.
type BitcoinTxInfo struct {
	// TxID reported by Bitcoin
	TxID    uint64          `json:"tx_id"`
	BtcTxID []byte          `json:"btc_tx_id"`
	Status  BitcoinTxStatus `json:"status"`
}

// IkaTxStatus represents the different states of a native transaction.
type IkaTxStatus byte

// Ika transaction status constants
const (
	Success IkaTxStatus = iota
	Failed
)

// BitcoinTxStatus represents the different states of a bitcoin transaction.
type BitcoinTxStatus byte

// Bitcoin transaction status constants
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
			timestamp INTEGER NOT NULL
		)`
	const createIkaTxsTableSQL = `
		CREATE TABLE IF NOT EXISTS ika_txs (
			sr_id INTEGER NOT NULL,  -- sign request
			status INTEGER NOT NULL,
			ika_tx_id TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			note TEXT,
			PRIMARY KEY (sr_id, ika_tx_id),
			FOREIGN KEY (sr_id) REFERENCES ika_sign_requests (id)
		)`

	const createBitcoinTxsTableSQL = `
		CREATE TABLE IF NOT EXISTS bitcoin_txs (
			sr_id INTEGER NOT NULL,
			status INTEGER NOT NULL,
			btc_tx_id BlOB NOT NULL,
			timestamp INTEGER NOT NULL,
			note TEXT,
			PRIMARY KEY (sr_id, btc_tx_id),
			FOREIGN KEY (sr_id) REFERENCES ika_sign_requests (id)
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
		INSERT INTO ika_sign_requests (id, payload, dwallet_id, user_sig, final_sig, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(
		insertIkaSignRequestSQL,
		signReq.ID,
		signReq.Payload,
		signReq.DWalletID,
		signReq.UserSig,
		signReq.FinalSig,
		signReq.Timestamp,
	)
	return err
}

// InsertIkaTx inserts a new Ika transaction into the database.
func (db *DB) InsertIkaTx(tx IkaTx) error {
	const insertIkaTxSQL = `
		INSERT INTO ika_txs (sr_id, status, ika_tx_id, timestamp, note)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertIkaTxSQL, tx.SrID, tx.Status, tx.IkaTxID, tx.Timestamp, tx.Note)
	return err
}

// InsertBtcTx inserts a new Bitcoin transaction into the database.
func (db *DB) InsertBtcTx(tx BitcoinTx) error {
	const insertBtcTxSQL = `
		INSERT INTO bitcoin_txs (sr_id, status, btc_tx_id, timestamp, note)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertBtcTxSQL, tx.SrID, tx.Status, tx.BtcTxID, tx.Timestamp, tx.Note)
	return err
}

// GetIkaSignRequestByID retrives a signature request by its id
func (db DB) GetIkaSignRequestByID(id uint64) (*IkaSignRequest, error) {
	const getIkaSignRequestByIDSQL = `
		SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
		FROM ika_sign_requests
		WHERE id = ?`
	row := db.conn.QueryRow(getIkaSignRequestByIDSQL, id)
	var signReq IkaSignRequest
	err := row.Scan(
		&signReq.ID,
		&signReq.Payload,
		&signReq.DWalletID,
		&signReq.UserSig,
		&signReq.FinalSig,
		&signReq.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &signReq, nil
}

// GetIkaTx retrieves an Ika transaction by its primary key (sr_id and ika_tx_id).
func (db *DB) GetIkaTx(signRequestID uint64, ikaTxID string) (*IkaTx, error) {
	const getIkaTxByTxIDAndIkaTxIDSQL = `
        SELECT sr_id, status, ika_tx_id, timestamp, note
        FROM ika_txs
        WHERE sr_id = ? AND ika_tx_id = ?`
	row := db.conn.QueryRow(getIkaTxByTxIDAndIkaTxIDSQL, signRequestID, ikaTxID)
	var ikaTx IkaTx
	err := row.Scan(&ikaTx.SrID, &ikaTx.Status, &ikaTx.IkaTxID, &ikaTx.Timestamp, &ikaTx.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // no rows
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &ikaTx, nil
}

// GetBitcoinTx retrieves a Bitcoin transaction by its primary key (sr_id and btc_tx_id).
func (db *DB) GetBitcoinTx(signRequestID uint64, btcTxID []byte) (*BitcoinTx, error) {
	const getBitcoinTxByTxIDAndBtcTxIDSQL = `
        SELECT sr_id, status, btc_tx_id, timestamp, note
        FROM bitcoin_txs
        WHERE sr_id = ? AND btc_tx_id = ?`
	row := db.conn.QueryRow(getBitcoinTxByTxIDAndBtcTxIDSQL, signRequestID, btcTxID)
	var bitcoinTx BitcoinTx
	err := row.Scan(&bitcoinTx.SrID, &bitcoinTx.Status, &bitcoinTx.BtcTxID, &bitcoinTx.Timestamp, &bitcoinTx.Note)
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
	const getIkaSignRequestWithStatusSQL = `
        SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp, it.status
        FROM ika_sign_requests sr
        INNER JOIN ika_txs it ON sr.id = it.sr_id
        WHERE sr.id = ?
        ORDER BY it.time DESC -- Get the latest status
        LIMIT 1`
	row := db.conn.QueryRow(getIkaSignRequestWithStatusSQL, id)
	var request IkaSignRequest
	var status IkaTxStatus
	err := row.Scan(
		&request.ID,
		&request.Payload,
		&request.DWalletID,
		&request.UserSig,
		&request.FinalSig,
		&request.Timestamp,
		&status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to query IkaSignRequest & IkaTxStatus: %w", err)
	}

	return &request, status, nil
}

// GetBitcoinTxsToBroadcast retrieves IkaSignRequests that have been signed by IKA
// and are due to be broadcasted to bitcoin.
//
// This function checks for the following conditions:
// - The IkaSignRequest must have a final signature (final_sig IS NOT NULL).
// - There must be no corresponding entry in the bitcoin_txs table, OR
// - There must be only one corresponding entry in the bitcoin_txs table with a status of "Pending".
//
// The reason for checking these conditions is that we cannot have a Bitcoin transaction hash (btc_tx_id)
// before the first broadcast attempt.
func (db *DB) GetBitcoinTxsToBroadcast() ([]IkaSignRequest, error) {
	const getSignedIkaSignRequestsSQL = `
        SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp
        FROM ika_sign_requests sr
        LEFT JOIN bitcoin_txs bt ON sr.id = bt.sr_id
        WHERE sr.final_sig IS NOT NULL
        GROUP BY sr.id
        HAVING COUNT(CASE WHEN bt.status = ? THEN 1 ELSE NULL END) = COUNT(bt.sr_id)`
	rows, err := db.conn.Query(getSignedIkaSignRequestsSQL, Pending)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()
	var requests []IkaSignRequest
	for rows.Next() {
		var request IkaSignRequest
		err := rows.Scan(
			&request.ID,
			&request.Payload,
			&request.DWalletID,
			&request.UserSig,
			&request.FinalSig,
			&request.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetBroadcastedBitcoinTxsInfo queries Bitcoin transactions that has been braodcasted but not confirmed.
// that do not have a "Confirmed" status.
func (db *DB) GetBroadcastedBitcoinTxsInfo() ([]BitcoinTxInfo, error) {
	const getBroadcastedBitcoinTxsInfoSQL = `
        SELECT bt.sr_id, bt.btc_tx_id, bt.status
        FROM bitcoin_txs bt
        WHERE bt.status = ?
        AND NOT EXISTS (
            SELECT 1
            FROM bitcoin_txs bt2
            WHERE bt2.sr_id = bt.sr_id AND bt2.status = ?
        )`
	rows, err := db.conn.Query(getBroadcastedBitcoinTxsInfoSQL, Broadcasted, Confirmed)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var txs []BitcoinTxInfo
	for rows.Next() {
		var tx BitcoinTxInfo
		err := rows.Scan(&tx.TxID, &tx.BtcTxID, &tx.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

// GetPendingIkaSignRequests retrieves IkaSignRequests that need to be signed.
func (db *DB) GetPendingIkaSignRequests() ([]IkaSignRequest, error) {
	const getPendingIkaSignRequestsSQL = `
        SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
        FROM ika_sign_requests
        WHERE final_sig IS NULL`
	rows, err := db.conn.Query(getPendingIkaSignRequestsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()
	var requests []IkaSignRequest
	for rows.Next() {
		var request IkaSignRequest
		err := rows.Scan(
			&request.ID,
			&request.Payload,
			&request.DWalletID,
			&request.UserSig,
			&request.FinalSig,
			&request.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// UpdateIkaSignRequestFinalSig updates the final signature of an IkaSignRequest in the database.
func (db *DB) UpdateIkaSignRequestFinalSig(id uint64, finalSig Signature) error {
	const updateIkaSignRequestFinalSigSQL = `
        UPDATE ika_sign_requests 
        SET final_sig = ?
        WHERE id = ?`
	_, err := db.conn.Exec(updateIkaSignRequestFinalSigSQL, finalSig, id)

	if err != nil {
		return fmt.Errorf("failed to update the final signature: %w", err)
	}

	return nil
}

// UpdateBitcoinTxToConfirmed updates the bitcoin transaction to `Confirmed`.
func (db *DB) UpdateBitcoinTxToConfirmed(id uint64, txID []byte) error {
	const updateBitcoinTxToConfirmedSQL = `
        UPDATE bitcoin_txs 
        SET status = ?, timestamp = ?
        WHERE sr_id = ? AND btc_tx_id = ?`
	timestamp := time.Now().Unix()
	_, err := db.conn.Exec(updateBitcoinTxToConfirmedSQL, Confirmed, timestamp, id, txID)

	if err != nil {
		return fmt.Errorf("failed to update the status : %w", err)
	}

	return nil
}

// Close closes the db connection
func (db DB) Close() error {
	return db.conn.Close()
}
