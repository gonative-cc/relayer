package clients

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

//TODO: this are the methods implemented in sui lc
// we need to provide implementations for it so it calls the sui smart contract method
// we can use the already implemented ika/sui client

// NOTE: not copied
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
