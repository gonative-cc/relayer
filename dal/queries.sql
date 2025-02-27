-- name: InsertIkaSignRequest :exec
INSERT INTO ika_sign_requests (id, payload, dwallet_id, user_sig, final_sig, timestamp) 
VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertIkaTx :exec
INSERT INTO ika_txs (sr_id, status, ika_tx_id, timestamp, note)
VALUES (?, ?, ?, ?, ?);

-- name: InsertBtcTx :exec
INSERT INTO bitcoin_txs (sr_id, status, btc_tx_id, timestamp, note)
VALUES (?, ?, ?, ?, ?);

-- name: GetIkaSignRequestByID :one
SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
FROM ika_sign_requests
WHERE id = ?;

-- name: GetIkaTx :one
SELECT sr_id, status, ika_tx_id, timestamp, note
FROM ika_txs
WHERE sr_id = ? AND ika_tx_id = ?;

-- name: GetBitcoinTx :one
SELECT sr_id, status, btc_tx_id, timestamp, note
FROM bitcoin_txs
WHERE sr_id = ? AND btc_tx_id = ?;

-- name: GetIkaSignRequestWithStatus :one
SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp, it.status
FROM ika_sign_requests sr
INNER JOIN ika_txs it ON sr.id = it.sr_id
WHERE sr.id = ?
ORDER BY it.timestamp DESC -- Get the latest status
LIMIT 1;

-- name: GetBitcoinTxsToBroadcast :many
SELECT sr.id, sr.payload, sr.dwallet_id, sr.user_sig, sr.final_sig, sr.timestamp
FROM ika_sign_requests sr
LEFT JOIN bitcoin_txs bt ON sr.id = bt.sr_id
WHERE sr.final_sig IS NOT NULL
GROUP BY sr.id
HAVING COUNT(CASE WHEN bt.status = ? THEN 1 ELSE NULL END) = COUNT(bt.sr_id);

-- name: GetBroadcastedBitcoinTxsInfo :many
SELECT bt.sr_id, bt.btc_tx_id, bt.status
FROM bitcoin_txs bt
WHERE bt.status = 1 -- `Broadcasted`
AND NOT EXISTS (
    SELECT 1
    FROM bitcoin_txs bt2
    WHERE bt2.sr_id = bt.sr_id AND bt2.status = 2 -- `Confirmed`
);

-- name: GetPendingIkaSignRequests :many
SELECT id, payload, dwallet_id, user_sig, final_sig, timestamp
FROM ika_sign_requests
WHERE final_sig IS NULL;

-- name: UpdateIkaSignRequestFinalSig :exec
UPDATE ika_sign_requests 
SET final_sig = ?
WHERE id = ?;

-- name: UpdateBitcoinTxToConfirmed :exec
UPDATE bitcoin_txs 
SET status = ?, timestamp = ?
WHERE sr_id = ? AND btc_tx_id = ?;