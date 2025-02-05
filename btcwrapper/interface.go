package btcwrapper

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	relayertypes "github.com/gonative-cc/relayer/bitcoinspv/types"
)

// BTCClient defines the interface for interacting with a Bitcoin node
type BTCClient interface {
	// Node lifecycle methods
	Stop()
	WaitForShutdown()

	// Block subscription methods
	SubscribeNewBlocks()
	BlockEventChannel() <-chan *relayertypes.BlockEvent

	// Block query methods
	GetBestBlock() (*chainhash.Hash, int64, error)
	GetBlockByHash(blockHash *chainhash.Hash) (*relayertypes.IndexedBlock, *wire.MsgBlock, error)
	GetTailBlocksByHeight(height int64) ([]*relayertypes.IndexedBlock, error)
	GetBlockByHeight(height int64) (*relayertypes.IndexedBlock, *wire.MsgBlock, error)

	// Transaction methods
	GetTxOut(txHash *chainhash.Hash, index uint32, mempool bool) (*btcjson.GetTxOutResult, error)
	GetTransaction(txHash *chainhash.Hash) (*btcjson.GetTransactionResult, error)
	GetRawTransaction(txHash *chainhash.Hash) (*btcutil.Tx, error)
}
