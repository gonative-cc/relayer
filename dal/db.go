package dal

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// IkaSignRequest represents a row in the `ika_sign_requests` table.
type IkaSignRequest struct {
	DWalletID string    `json:"dwallet_id"`
	UserSig   string    `json:"user_sig"`
	Payload   Payload   `json:"payload"`
	FinalSig  Signature `json:"final_sig"`
	ID        uint64    `json:"id"` // ID is a sign request reported and managed by Native
	Timestamp int64     `json:"time"`
}

// IkaTx represents a row in the `ika_txs` table.
type IkaTx struct {
	IkaTxID   string      `json:"ika_tx_id"`
	Note      string      `json:"note"`
	SrID      uint64      `json:"sr_id"` // SrID is the IkaSignRequest.ID (from Native)
	Timestamp int64       `json:"timestamp"`
	Status    IkaTxStatus `json:"status"`
}

// BitcoinTx represents a row in the `bitcoin_txs` table.
type BitcoinTx struct {
	Note      string          `json:"note"`
	BtcTxID   []byte          `json:"btc_tx_id"`
	SrID      uint64          `json:"sr_id"` // SrID is the IkaSignRequest.ID (from Native)
	Timestamp int64           `json:"timestamp"`
	Status    BitcoinTxStatus `json:"status"`
}

// BitcoinTxInfo holds the relevant information for a Bitcoin transaction.
type BitcoinTxInfo struct {
	BtcTxID []byte          `json:"btc_tx_id"`
	TxID    uint64          `json:"tx_id"` // TxID reported by Bitcoin
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
	conn  *sql.DB
	mutex *sync.RWMutex
}

// NewDB creates a new DB instance
func NewDB(dbPath string) (DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return DB{}, fmt.Errorf("dal: can't open sqlite3: %w", err)
	}
	return DB{conn: conn, mutex: &sync.RWMutex{}}, err
}

// InitDB initializes the database and creates the tables.
func (db DB) InitDB() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

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
		return fmt.Errorf("dal: creating ika_sign_requests table: %w", err)
	}
	_, err = db.conn.Exec(createIkaTxsTableSQL)
	if err != nil {
		return fmt.Errorf("dal: creating ika_txs table: %w", err)
	}
	_, err = db.conn.Exec(createBitcoinTxsTableSQL)
	if err != nil {
		return fmt.Errorf("dal: creating bitcoin_txs table: %w", err)
	}
	return nil
}

// InsertIkaSignRequest inserts a new transaction into the database
func (db DB) InsertIkaSignRequest(signReq IkaSignRequest) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

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
	if err != nil {
		return fmt.Errorf("dal: inserting ika_sign_request: %w", err)
	}
	return nil
}

// InsertIkaTx inserts a new Ika transaction into the database.
func (db DB) InsertIkaTx(tx IkaTx) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	const insertIkaTxSQL = `
		INSERT INTO ika_txs (sr_id, status, ika_tx_id, timestamp, note)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertIkaTxSQL, tx.SrID, tx.Status, tx.IkaTxID, tx.Timestamp, tx.Note)
	if err != nil {
		return fmt.Errorf("dal: inserting ika_tx: %w", err)
	}
	return nil
}

// InsertBtcTx inserts a new Bitcoin transaction into the database.
func (db DB) InsertBtcTx(tx BitcoinTx) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	const insertBtcTxSQL = `
		INSERT INTO bitcoin_txs (sr_id, status, btc_tx_id, timestamp, note)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(insertBtcTxSQL, tx.SrID, tx.Status, tx.BtcTxID, tx.Timestamp, tx.Note)
	if err != nil {
		return fmt.Errorf("dal: inserting bitcoin_tx: %w", err)
	}
	return nil
}

// GetIkaSignRequestByID retrives a signature request by its id
func (db DB) GetIkaSignRequestByID(id uint64) (*IkaSignRequest, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

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
		return nil, fmt.Errorf("dal: getting ika_sign_request by id: %w", err)
	}
	return &signReq, nil
}

// GetIkaTx retrieves an Ika transaction by its primary key (sr_id and ika_tx_id).
func (db DB) GetIkaTx(signRequestID uint64, ikaTxID string) (*IkaTx, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	const getIkaTxByTxIDAndIkaTxIDSQL = `
        SELECT sr_id, status, ika_tx_id, timestamp, note
        FROM ika_txs
        WHERE sr_id = ? AND ika_tx_id = ?`
	row := db.conn.QueryRow(getIkaTxByTxIDAndIkaTxIDSQL, signRequestID, ikaTxID)
	var ikaTx IkaTx
	err := row.Scan(&ikaTx.SrID, &ikaTx.Status, &ikaTx.IkaTxID, &ikaTx.Timestamp, &ikaTx.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("dal: getting ika_tx: %w", err)
	}
	return &ikaTx, nil
}

// GetBitcoinTx retrieves a Bitcoin transaction by its primary key (sr_id and btc_tx_id).
func (db DB) GetBitcoinTx(signRequestID uint64, btcTxID []byte) (*BitcoinTx, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	const getBitcoinTxByTxIDAndBtcTxIDSQL = `
        SELECT sr_id, status, btc_tx_id, timestamp, note
        FROM bitcoin_txs
        WHERE sr_id = ? AND btc_tx_id = ?`
	row := db.conn.QueryRow(getBitcoinTxByTxIDAndBtcTxIDSQL, signRequestID, btcTxID)
	var bitcoinTx BitcoinTx
	err := row.Scan(&bitcoinTx.SrID, &bitcoinTx.Status, &bitcoinTx.BtcTxID, &bitcoinTx.Timestamp, &bitcoinTx.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("dal: getting bitcoin_tx: %w", err)
	}
	return &bitcoinTx, nil
}

// GetIkaSignRequestWithStatus retrieves an IkaSignRequest with its associated IkaTx status.
func (db DB) GetIkaSignRequestWithStatus(id uint64) (*IkaSignRequest, IkaTxStatus, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

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
		return nil, 0, fmt.Errorf("dal: getting ika_sign_request with status: %w", err)
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
func (db DB) GetBitcoinTxsToBroadcast() ([]IkaSignRequest, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	const getSignedIkaSignRequestsSQL = `
        SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp
        FROM ika_sign_requests sr
        LEFT JOIN bitcoin_txs bt ON sr.id = bt.sr_id
        WHERE sr.final_sig IS NOT NULL
        GROUP BY sr.id
        HAVING COUNT(CASE WHEN bt.status = ? THEN 1 ELSE NULL END) = COUNT(bt.sr_id)`
	rows, err := db.conn.Query(getSignedIkaSignRequestsSQL, Pending)
	if err != nil {
		return nil, fmt.Errorf("dal: querying bitcoin_txs to broadcast: %w", err)
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
			return nil, fmt.Errorf("dal: scanning row for bitcoin_txs to broadcast: %w", err)
		}
		requests = append(requests, request)
	}
	return requests, nil
}

// GetBroadcastedBitcoinTxsInfo queries Bitcoin transactions that has been braodcasted but not confirmed.
// that do not have a "Confirmed" status.
func (db DB) GetBroadcastedBitcoinTxsInfo() ([]BitcoinTxInfo, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

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
		return nil, fmt.Errorf("dal: querying broadcasted bitcoin_txs info: %w", err)
	}
	defer rows.Close()

	var txs []BitcoinTxInfo
	for rows.Next() {
		var tx BitcoinTxInfo
		err := rows.Scan(&tx.TxID, &tx.BtcTxID, &tx.Status)
		if err != nil {
			return nil, fmt.Errorf("dal: scanning row for broadcasted bitcoin_txs info: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// GetPendingIkaSignRequests retrieves IkaSignRequests that need to be signed.
func (db DB) GetPendingIkaSignRequests() ([]IkaSignRequest, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	const getPendingIkaSignRequestsSQL = `
        SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
        FROM ika_sign_requests
        WHERE final_sig IS NULL`
	rows, err := db.conn.Query(getPendingIkaSignRequestsSQL)
	if err != nil {
		return nil, fmt.Errorf("dal: querying pending ika_sign_requests: %w", err)
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
			return nil, fmt.Errorf("dal: scanning row for pending ika_sign_requests: %w", err)
		}
		requests = append(requests, request)
	}
	return requests, nil
}

// UpdateIkaSignRequestFinalSig updates the final signature of an IkaSignRequest in the database.
func (db DB) UpdateIkaSignRequestFinalSig(id uint64, finalSig Signature) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	const updateIkaSignRequestFinalSigSQL = `
        UPDATE ika_sign_requests 
        SET final_sig = ?
        WHERE id = ?`
	_, err := db.conn.Exec(updateIkaSignRequestFinalSigSQL, finalSig, id)

	if err != nil {
		return fmt.Errorf("dal: updating ika_sign_request final sig: %w", err)
	}
	return nil
}

// UpdateBitcoinTxToConfirmed updates the bitcoin transaction to `Confirmed`.
func (db DB) UpdateBitcoinTxToConfirmed(id uint64, txID []byte) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	const updateBitcoinTxToConfirmedSQL = `
        UPDATE bitcoin_txs 
        SET status = ?, timestamp = ?
        WHERE sr_id = ? AND btc_tx_id = ?`
	timestamp := time.Now().Unix()
	_, err := db.conn.Exec(updateBitcoinTxToConfirmedSQL, Confirmed, timestamp, id, txID)

	if err != nil {
		return fmt.Errorf("dal: updating bitcoin_tx to confirmed: %w", err)
	}
	return nil
}

// Close closes the db connection
func (db DB) Close() error {
	// make sure other read / writes are done
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.conn.Close()
}
