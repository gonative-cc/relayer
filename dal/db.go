package dal

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

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
	Querier
}

// NewDB creates a new DB instance
func NewDB(dbPath string) (DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return DB{}, fmt.Errorf("dal: can't open sqlite3: %w", err)
	}
	queries := New(conn)
	return DB{conn: conn, mutex: &sync.RWMutex{}, Querier: queries}, err
}

//go:embed schema.sql
var content embed.FS

// InitDB initializes the database and creates the tables.
func (db DB) InitDB() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	ctx := context.Background()
	schema, err := content.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("dal: reading schema file: %w", err)
	}
	_, err = db.conn.ExecContext(ctx, string(schema))
	if err != nil {
		return fmt.Errorf("dal: creating tables: %w", err)
	}
	return nil
}

// InsertIkaSignRequest inserts a new transaction into the database
func (db DB) InsertIkaSignRequest(ctx context.Context, signReq IkaSignRequest) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	params := InsertIkaSignRequestParams(signReq)
	err := db.Querier.InsertIkaSignRequest(ctx, &params)
	if err != nil {
		return fmt.Errorf("dal: inserting ika_sign_request: %w", err)
	}
	return nil
}

// InsertIkaTx inserts a new Ika transaction into the database.
func (db DB) InsertIkaTx(ctx context.Context, tx IkaTx) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	params := InsertIkaTxParams(tx)
	err := db.Querier.InsertIkaTx(ctx, &params)
	if err != nil {
		return fmt.Errorf("dal: inserting ika_tx: %w", err)
	}
	return nil
}

// InsertBtcTx inserts a new Bitcoin transaction into the database.
func (db DB) InsertBtcTx(ctx context.Context, tx BitcoinTx) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	params := InsertBtcTxParams(tx)
	err := db.Querier.InsertBtcTx(ctx, &params)
	if err != nil {
		return fmt.Errorf("dal: inserting bitcoin_tx: %w", err)
	}
	return nil
}

// GetIkaSignRequestByID retrives a signature request by its id
func (db DB) GetIkaSignRequestByID(ctx context.Context, id int64) (*IkaSignRequest, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	signReq, err := db.Querier.GetIkaSignRequestByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows { // TODO test if we can just return it since we are using `emit_empty_slices: true` flag
			return nil, nil // Return nil, nil for not found
		}
		return nil, fmt.Errorf("dal: getting ika_sign_request by id: %w", err)
	}
	return signReq, nil
}

// GetIkaTx retrieves an Ika transaction by its primary key (sr_id and ika_tx_id).
func (db DB) GetIkaTx(ctx context.Context, signRequestID int64, ikaTxID string) (*IkaTx, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	ikaTx, err := db.Querier.GetIkaTx(ctx, &GetIkaTxParams{SrID: signRequestID, IkaTxID: ikaTxID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("dal: getting ika_tx: %w", err)
	}
	return ikaTx, nil
}

// GetBitcoinTx retrieves a Bitcoin transaction by its primary key (sr_id and btc_tx_id).
func (db DB) GetBitcoinTx(ctx context.Context, signRequestID int64, btcTxID []byte) (*BitcoinTx, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	bitcoinTx, err := db.Querier.GetBitcoinTx(ctx, &GetBitcoinTxParams{SrID: signRequestID, BtcTxID: btcTxID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("dal: getting bitcoin_tx: %w", err)
	}
	return bitcoinTx, nil
}

// GetIkaSignRequestWithStatus retrieves an IkaSignRequest with its associated IkaTx status.
func (db DB) GetIkaSignRequestWithStatus(ctx context.Context, id int64) (*GetIkaSignRequestWithStatusRow, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	result, err := db.Querier.GetIkaSignRequestWithStatus(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("dal: getting ika_sign_request with status: %w", err)
	}
	return result, nil // Return the result and converted status
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
func (db DB) GetBitcoinTxsToBroadcast(ctx context.Context) ([]*IkaSignRequest, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	requests, err := db.Querier.GetBitcoinTxsToBroadcast(ctx, int64(Pending))
	if err != nil {
		return nil, fmt.Errorf("dal: querying bitcoin_txs to broadcast: %w", err)
	}
	return requests, nil
}

// GetBroadcastedBitcoinTxsInfo queries Bitcoin transactions that has been braodcasted but not confirmed.
// that do not have a "Confirmed" status.
func (db DB) GetBroadcastedBitcoinTxsInfo(ctx context.Context) ([]*GetBroadcastedBitcoinTxsInfoRow, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	txs, err := db.Querier.GetBroadcastedBitcoinTxsInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("dal: querying broadcasted bitcoin_txs info: %w", err)
	}
	return txs, nil
}

// GetPendingIkaSignRequests retrieves IkaSignRequests that need to be signed.
func (db DB) GetPendingIkaSignRequests(ctx context.Context) ([]*IkaSignRequest, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	requests, err := db.Querier.GetPendingIkaSignRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("dal: querying pending ika_sign_requests: %w", err)
	}
	return requests, nil
}

// UpdateIkaSignRequestFinalSig updates the final signature of an IkaSignRequest in the database.
func (db DB) UpdateIkaSignRequestFinalSig(ctx context.Context, id int64, finalSig Signature) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.Querier.UpdateIkaSignRequestFinalSig(
		ctx,
		&UpdateIkaSignRequestFinalSigParams{ID: id, FinalSig: finalSig},
	)
	if err != nil {
		return fmt.Errorf("dal: updating ika_sign_request final sig: %w", err)
	}
	return nil
}

// UpdateBitcoinTxToConfirmed updates the bitcoin transaction to `Confirmed`.
func (db DB) UpdateBitcoinTxToConfirmed(ctx context.Context, id int64, txID []byte) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	timestamp := time.Now().Unix()

	err := db.Querier.UpdateBitcoinTxToConfirmed(
		ctx,
		&UpdateBitcoinTxToConfirmedParams{SrID: id, BtcTxID: txID, Status: int64(Confirmed), Timestamp: timestamp},
	)
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
