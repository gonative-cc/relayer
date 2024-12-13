package reporter

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type NativeClient interface {
	// txn to insert bitcoin block headers to native light client
	InsertHeaders(blockHeaders []*wire.BlockHeader) error
	// returns if given block hash is already written to native light client
	ContainsBTCBlock(blockHash *chainhash.Hash) (bool, error)
	// returns the block height and hash of tip block stored in native light client
	BTCHeaderChainTip() (*chainhash.Hash, error)
	Stop() error
}
