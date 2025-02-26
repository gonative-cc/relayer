// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package dal

import (
	"database/sql"
)

type BitcoinTx struct {
	SrID      int64           `json:"sr_id"`
	Status    BitcoinTxStatus `json:"status"`
	BtcTxID   []byte          `json:"btc_tx_id"`
	Timestamp int64           `json:"timestamp"`
	Note      sql.NullString  `json:"note"`
}

type IkaSignRequest struct {
	ID        int64  `json:"id"`
	Payload   []byte `json:"payload"`
	DWalletID string `json:"dwallet_id"`
	UserSig   string `json:"user_sig"`
	FinalSig  []byte `json:"final_sig"`
	Timestamp int64  `json:"timestamp"`
}

type IkaTx struct {
	SrID      int64          `json:"sr_id"`
	Status    IkaTxStatus    `json:"status"`
	IkaTxID   string         `json:"ika_tx_id"`
	Timestamp int64          `json:"timestamp"`
	Note      sql.NullString `json:"note"`
}
