package bitcoin

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Client defines the interface with only the functions we need.
type Client interface {
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
	Shutdown()
	GetTransaction(txHash *chainhash.Hash) (*btcjson.GetTransactionResult, error)
}
