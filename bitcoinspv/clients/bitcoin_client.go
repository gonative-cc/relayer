package clients

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/gonative-cc/relayer/bitcoinspv/types"

	btctypes "github.com/gonative-cc/relayer/bitcoinspv/types/btc"
)

// BTCClient is an abstraction over bitcoin node implementations (bitcoind, btcd).
// Refer to btcwrapper/ dir for implementation.
type BTCClient interface {
	Stop()
	WaitForShutdown()
	SubscribeNewBlocks()
	BlockEventChannel() <-chan *btctypes.BlockEvent
	GetBTCTipBlock() (*chainhash.Hash, uint64, error)
	// TODO: lets consider removing the wire.MsgBlock;
	// instead lets remove transactions from IndexedBlocks, rename it to: lightBLock and use just this
	GetBTCBlockByHash(blockHash *chainhash.Hash) (*types.IndexedBlock, *wire.MsgBlock, error)
	GetBTCTailBlocksByHeight(height uint64, fullBlocks bool) ([]*types.IndexedBlock, error)
	GetBTCBlockByHeight(height uint64) (*types.IndexedBlock, *wire.MsgBlock, error)
	GetBTCBlockHeaderByHeight(height uint64) (*wire.BlockHeader, error)
}
