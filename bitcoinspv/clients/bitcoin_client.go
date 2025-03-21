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
	GetBTCTipBlock() (*chainhash.Hash, int64, error)
	GetBTCBlockByHash(blockHash *chainhash.Hash) (*types.IndexedBlock, *wire.MsgBlock, error)
	GetBTCTailBlocksByHeight(height int64) ([]*types.IndexedBlock, error)
	GetBTCBlockByHeight(height int64) (*types.IndexedBlock, *wire.MsgBlock, error)
	GetBTCTailBlockHeadersByHeight(baseHeight int64) ([]*wire.BlockHeader, error)
	GetBTCBlockHeaderByHeight(height int64) (*wire.BlockHeader, error)
}
