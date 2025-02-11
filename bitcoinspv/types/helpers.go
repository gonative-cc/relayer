package types

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

// GetWrappedTxs returns transactions from a block message
func GetWrappedTxs(msg *wire.MsgBlock) []*btcutil.Tx {
	btcTxs := make([]*btcutil.Tx, 0, len(msg.Transactions))

	for i, tx := range msg.Transactions {
		newTx := btcutil.NewTx(tx)
		newTx.SetIndex(i)

		btcTxs = append(btcTxs, newTx)
	}

	return btcTxs
}
