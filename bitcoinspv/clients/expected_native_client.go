package clients

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// NativeClient is interface for interacting with bitcoin-lightclient.
// Refer to github.com/gonative-cc/bitcoin-lightclient for implementation.
type NativeClient interface {
	// txn to insert bitcoin block headers to native light client
	InsertHeaders(blockHeaders []*wire.BlockHeader) error
	// returns if given block hash is already written to native light client
	ContainsBTCBlock(blockHash *chainhash.Hash) (bool, error)
	// returns the block height and hash of tip block stored in native light client
	BTCHeaderChainTip() (int64, *chainhash.Hash, error)
	// returns if spvProof is valid or not
	VerifySPV(spvProof types.SPVProof) (int, error)
	Stop() error
}
