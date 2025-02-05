package clients

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/gonative-cc/relayer/bitcoinspv/types"
)

// BTCClient is an abstraction over bitcoin node implementations (bitcoind, btcd).
// Refer to btcwrapper/ dir for implementation.
type BTCClient interface {
	Stop()
	WaitForShutdown()
	SubscribeNewBlocks()
	BlockEventChannel() <-chan *types.BlockEvent
	GetBestBlock() (*chainhash.Hash, int64, error)
	GetBlockByHash(blockHash *chainhash.Hash) (*types.IndexedBlock, *wire.MsgBlock, error)
	GetTailBlocksByHeight(height int64) ([]*types.IndexedBlock, error)
	GetBlockByHeight(height int64) (*types.IndexedBlock, *wire.MsgBlock, error)
	GetTxOut(txHash *chainhash.Hash, index uint32, mempool bool) (*btcjson.GetTxOutResult, error)
	GetTransaction(txHash *chainhash.Hash) (*btcjson.GetTransactionResult, error)
	GetRawTransaction(txHash *chainhash.Hash) (*btcutil.Tx, error)
}
