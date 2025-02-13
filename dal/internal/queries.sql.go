// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package internal

import (
	"context"
	"database/sql"
)

const getBitcoinTx = `-- name: GetBitcoinTx :one
SELECT sr_id, status, btc_tx_id, timestamp, note
FROM bitcoin_txs
WHERE sr_id = ? AND btc_tx_id = ?
`

type GetBitcoinTxParams struct {
	SrID    int64  `json:"sr_id"`
	BtcTxID []byte `json:"btc_tx_id"`
}

func (q *Queries) GetBitcoinTx(ctx context.Context, arg *GetBitcoinTxParams) (*BitcoinTx, error) {
	row := q.db.QueryRowContext(ctx, getBitcoinTx, arg.SrID, arg.BtcTxID)
	var i BitcoinTx
	err := row.Scan(
		&i.SrID,
		&i.Status,
		&i.BtcTxID,
		&i.Timestamp,
		&i.Note,
	)
	return &i, err
}

const getBitcoinTxsToBroadcast = `-- name: GetBitcoinTxsToBroadcast :many
SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp
FROM ika_sign_requests sr
LEFT JOIN bitcoin_txs bt ON sr.id = bt.sr_id
WHERE sr.final_sig IS NOT NULL
GROUP BY sr.id
HAVING COUNT(CASE WHEN bt.status = ? THEN 1 ELSE NULL END) = COUNT(bt.sr_id)
`

func (q *Queries) GetBitcoinTxsToBroadcast(ctx context.Context, status int64) ([]*IkaSignRequest, error) {
	rows, err := q.db.QueryContext(ctx, getBitcoinTxsToBroadcast, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*IkaSignRequest{}
	for rows.Next() {
		var i IkaSignRequest
		if err := rows.Scan(
			&i.ID,
			&i.Payload,
			&i.DwalletID,
			&i.UserSig,
			&i.FinalSig,
			&i.Timestamp,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBroadcastedBitcoinTxsInfo = `-- name: GetBroadcastedBitcoinTxsInfo :many
SELECT bt.sr_id, bt.btc_tx_id, bt.status
FROM bitcoin_txs bt
WHERE bt.status = 1 -- ` + "`" + `Broadcasted` + "`" + `
AND NOT EXISTS (
    SELECT 1
    FROM bitcoin_txs bt2
    WHERE bt2.sr_id = bt.sr_id AND bt2.status = 2 -- ` + "`" + `Confirmed` + "`" + `
)
`

type GetBroadcastedBitcoinTxsInfoRow struct {
	SrID    int64  `json:"sr_id"`
	BtcTxID []byte `json:"btc_tx_id"`
	Status  int64  `json:"status"`
}

func (q *Queries) GetBroadcastedBitcoinTxsInfo(ctx context.Context) ([]*GetBroadcastedBitcoinTxsInfoRow, error) {
	rows, err := q.db.QueryContext(ctx, getBroadcastedBitcoinTxsInfo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*GetBroadcastedBitcoinTxsInfoRow{}
	for rows.Next() {
		var i GetBroadcastedBitcoinTxsInfoRow
		if err := rows.Scan(&i.SrID, &i.BtcTxID, &i.Status); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIkaSignRequestByID = `-- name: GetIkaSignRequestByID :one
SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
FROM ika_sign_requests
WHERE id = ?
`

func (q *Queries) GetIkaSignRequestByID(ctx context.Context, id int64) (*IkaSignRequest, error) {
	row := q.db.QueryRowContext(ctx, getIkaSignRequestByID, id)
	var i IkaSignRequest
	err := row.Scan(
		&i.ID,
		&i.Payload,
		&i.DwalletID,
		&i.UserSig,
		&i.FinalSig,
		&i.Timestamp,
	)
	return &i, err
}

const getIkaSignRequestWithStatus = `-- name: GetIkaSignRequestWithStatus :one
SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp, it.status
FROM ika_sign_requests sr
INNER JOIN ika_txs it ON sr.id = it.sr_id
WHERE sr.id = ?
ORDER BY it.timestamp DESC -- Get the latest status
LIMIT 1
`

type GetIkaSignRequestWithStatusRow struct {
	ID        int64  `json:"id"`
	Payload   []byte `json:"payload"`
	DwalletID string `json:"dwallet_id"`
	UserSig   string `json:"user_sig"`
	FinalSig  []byte `json:"final_sig"`
	Timestamp int64  `json:"timestamp"`
	Status    int64  `json:"status"`
}

func (q *Queries) GetIkaSignRequestWithStatus(ctx context.Context, id int64) (*GetIkaSignRequestWithStatusRow, error) {
	row := q.db.QueryRowContext(ctx, getIkaSignRequestWithStatus, id)
	var i GetIkaSignRequestWithStatusRow
	err := row.Scan(
		&i.ID,
		&i.Payload,
		&i.DwalletID,
		&i.UserSig,
		&i.FinalSig,
		&i.Timestamp,
		&i.Status,
	)
	return &i, err
}

const getIkaTx = `-- name: GetIkaTx :one
SELECT sr_id, status, ika_tx_id, timestamp, note
FROM ika_txs
WHERE sr_id = ? AND ika_tx_id = ?
`

type GetIkaTxParams struct {
	SrID    int64  `json:"sr_id"`
	IkaTxID string `json:"ika_tx_id"`
}

func (q *Queries) GetIkaTx(ctx context.Context, arg *GetIkaTxParams) (*IkaTx, error) {
	row := q.db.QueryRowContext(ctx, getIkaTx, arg.SrID, arg.IkaTxID)
	var i IkaTx
	err := row.Scan(
		&i.SrID,
		&i.Status,
		&i.IkaTxID,
		&i.Timestamp,
		&i.Note,
	)
	return &i, err
}

const getPendingIkaSignRequests = `-- name: GetPendingIkaSignRequests :many
SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
FROM ika_sign_requests
WHERE final_sig IS NULL
`

func (q *Queries) GetPendingIkaSignRequests(ctx context.Context) ([]*IkaSignRequest, error) {
	rows, err := q.db.QueryContext(ctx, getPendingIkaSignRequests)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*IkaSignRequest{}
	for rows.Next() {
		var i IkaSignRequest
		if err := rows.Scan(
			&i.ID,
			&i.Payload,
			&i.DwalletID,
			&i.UserSig,
			&i.FinalSig,
			&i.Timestamp,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertBtcTx = `-- name: InsertBtcTx :exec
INSERT INTO bitcoin_txs (sr_id, status, btc_tx_id, timestamp, note)
VALUES (?, ?, ?, ?, ?)
`

type InsertBtcTxParams struct {
	SrID      int64          `json:"sr_id"`
	Status    int64          `json:"status"`
	BtcTxID   []byte         `json:"btc_tx_id"`
	Timestamp int64          `json:"timestamp"`
	Note      sql.NullString `json:"note"`
}

func (q *Queries) InsertBtcTx(ctx context.Context, arg *InsertBtcTxParams) error {
	_, err := q.db.ExecContext(ctx, insertBtcTx,
		arg.SrID,
		arg.Status,
		arg.BtcTxID,
		arg.Timestamp,
		arg.Note,
	)
	return err
}

const insertIkaSignRequest = `-- name: InsertIkaSignRequest :exec
INSERT INTO ika_sign_requests (id, payload, dwallet_id, user_sig, final_sig, timestamp) 
VALUES (?, ?, ?, ?, ?, ?)
`

type InsertIkaSignRequestParams struct {
	ID        int64  `json:"id"`
	Payload   []byte `json:"payload"`
	DwalletID string `json:"dwallet_id"`
	UserSig   string `json:"user_sig"`
	FinalSig  []byte `json:"final_sig"`
	Timestamp int64  `json:"timestamp"`
}

func (q *Queries) InsertIkaSignRequest(ctx context.Context, arg *InsertIkaSignRequestParams) error {
	_, err := q.db.ExecContext(ctx, insertIkaSignRequest,
		arg.ID,
		arg.Payload,
		arg.DwalletID,
		arg.UserSig,
		arg.FinalSig,
		arg.Timestamp,
	)
	return err
}

const insertIkaTx = `-- name: InsertIkaTx :exec
INSERT INTO ika_txs (sr_id, status, ika_tx_id, timestamp, note)
VALUES (?, ?, ?, ?, ?)
`

type InsertIkaTxParams struct {
	SrID      int64          `json:"sr_id"`
	Status    int64          `json:"status"`
	IkaTxID   string         `json:"ika_tx_id"`
	Timestamp int64          `json:"timestamp"`
	Note      sql.NullString `json:"note"`
}

func (q *Queries) InsertIkaTx(ctx context.Context, arg *InsertIkaTxParams) error {
	_, err := q.db.ExecContext(ctx, insertIkaTx,
		arg.SrID,
		arg.Status,
		arg.IkaTxID,
		arg.Timestamp,
		arg.Note,
	)
	return err
}

const updateBitcoinTxToConfirmed = `-- name: UpdateBitcoinTxToConfirmed :exec
UPDATE bitcoin_txs 
SET status = ?, timestamp = ?
WHERE sr_id = ? AND btc_tx_id = ?
`

type UpdateBitcoinTxToConfirmedParams struct {
	Status    int64  `json:"status"`
	Timestamp int64  `json:"timestamp"`
	SrID      int64  `json:"sr_id"`
	BtcTxID   []byte `json:"btc_tx_id"`
}

func (q *Queries) UpdateBitcoinTxToConfirmed(ctx context.Context, arg *UpdateBitcoinTxToConfirmedParams) error {
	_, err := q.db.ExecContext(ctx, updateBitcoinTxToConfirmed,
		arg.Status,
		arg.Timestamp,
		arg.SrID,
		arg.BtcTxID,
	)
	return err
}

const updateIkaSignRequestFinalSig = `-- name: UpdateIkaSignRequestFinalSig :exec
UPDATE ika_sign_requests 
SET final_sig = ?
WHERE id = ?
`

type UpdateIkaSignRequestFinalSigParams struct {
	FinalSig []byte `json:"final_sig"`
	ID       int64  `json:"id"`
}

func (q *Queries) UpdateIkaSignRequestFinalSig(ctx context.Context, arg *UpdateIkaSignRequestFinalSigParams) error {
	_, err := q.db.ExecContext(ctx, updateIkaSignRequestFinalSig, arg.FinalSig, arg.ID)
	return err
}
