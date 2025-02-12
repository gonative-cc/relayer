package types

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

// GetWrappedTxs returns transactions from a block message
func GetWrappedTxs(msg *wire.MsgBlock) []*btcutil.Tx {
	txCount := len(msg.Transactions)
	wrappedTxs := make([]*btcutil.Tx, txCount)

	for idx := 0; idx < txCount; idx++ {
		tx := msg.Transactions[idx]
		wrappedTx := btcutil.NewTx(tx)
		wrappedTx.SetIndex(idx)
		wrappedTxs[idx] = wrappedTx
	}

	return wrappedTxs
}
